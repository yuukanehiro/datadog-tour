# Middlewareチェーンの仕組みとRecovery Middlewareの動作原理

## 目次
1. [Middlewareチェーンの基本構造](#middlewareチェーンの基本構造)
2. [なぜRecovery MiddlewareがすべてのPanicをキャッチできるのか](#なぜrecovery-middlewareがすべてのpanicをキャッチできるのか)
3. [実装例とコード解説](#実装例とコード解説)
4. [実行順序の詳細](#実行順序の詳細)
5. [Goの`defer`と`recover`の仕組み](#goのdeferとrecoverの仕組み)
6. [ベストプラクティス](#ベストプラクティス)

---

## Middlewareチェーンの基本構造

Echoなどのウェブフレームワークのミドルウェアは、**玉ねぎ構造（Onion Structure）** または **ロシア人形構造（Matryoshka Structure）** と呼ばれるパターンで動作します。

### 概念図

```
         ┌─────────────────────────────────────────────┐
         │  Recovery Middleware                        │
         │  ┌──────────────────────────────────────┐   │
         │  │  Logger Middleware                   │   │
         │  │  ┌───────────────────────────────┐   │   │
         │  │  │  CORS Middleware              │   │   │
         │  │  │  ┌────────────────────────┐   │   │   │
         │  │  │  │  Handler                │   │   │   │
         │  │  │  │  ┌─────────────────┐   │   │   │   │
         │  │  │  │  │  UseCase        │   │   │   │   │
         │  │  │  │  │  ┌──────────┐   │   │   │   │   │
         │  │  │  │  │  │Repository│   │   │   │   │   │
Request ─┼─>│─>│─>│─>│─>│  Panic!  │   │   │   │   │   │
         │  │  │  │  │  └──────────┘   │   │   │   │   │
         │  │  │  │  └─────────────────┘   │   │   │   │
         │  │  │  └────────────────────────┘   │   │   │
         │  │  └───────────────────────────────┘   │   │
         │  └──────────────────────────────────────┘   │
         │  ← ここでrecover()がPanicをキャッチ          │
         └─────────────────────────────────────────────┘
```

### リクエストの流れ

1. **リクエスト到着**: Recovery Middleware → Logger → CORS → Handler → UseCase → Repository
2. **レスポンス返却**: Repository → UseCase → Handler → CORS → Logger → Recovery Middleware
3. **Panic発生時**: どこでPanicが起きても、スタックが巻き戻り、最初のRecovery Middlewareでキャッチされる

---

## なぜRecovery MiddlewareがすべてのPanicをキャッチできるのか

### 質問

> middlewareの処理が終わって、presenter層のhandler→usecase→repositoryに処理が入って、なぜrecoveryMiddlewareのrecover()が効くの？

### 回答

これはGoの**ミドルウェアチェーンの「玉ねぎ構造」**と**`defer`/`recover`の仕組み**によるものです。

### キーポイント

1. **`defer func()`でrecover()を設定** (実行前に登録だけする)
2. **その後に`next(c)`を呼び出し** (後続の全処理が実行される)
3. **`next(c)`の中で実行される全ての処理は同じゴルーチン、同じ関数スタック内**
4. **Panicが発生してスタックが巻き戻ると、最初に設定した`defer`が実行される**

---

## 実装例とコード解説

### Recovery Middlewareの実装

```go
// internal/presentation/middleware/echo_recovery.go
package middleware

import (
    "fmt"
    "runtime/debug"

    "github.com/labstack/echo/v4"
    "gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

func EchoRecoveryMiddleware() echo.MiddlewareFunc {
    return func(next echo.HandlerFunc) echo.HandlerFunc {
        return func(c echo.Context) error {
            // 1. Spanを作成
            span, ctx := tracer.StartSpanFromContext(c.Request().Context(), "middleware.recovery")
            defer span.Finish()

            // 2. Contextを更新
            c.SetRequest(c.Request().WithContext(ctx))

            // 3. Trace情報を事前に抽出（Panic後はアクセスできないため）
            var traceID, spanID string
            if span != nil {
                spanContext := span.Context()
                traceID = strconv.FormatUint(spanContext.TraceID(), 10)
                spanID = strconv.FormatUint(spanContext.SpanID(), 10)
            }

            // 4. ★重要★ ここでrecover()を設定（まだ実行されない）
            defer func() {
                if err := recover(); err != nil {
                    // Panic発生時の処理
                    stackTrace := string(debug.Stack())

                    logger.WithFields(logrus.Fields{
                        "panic.value":       err,
                        "panic.stack_trace": stackTrace,
                        "dd.trace_id":       traceID,
                        "dd.span_id":        spanID,
                    }).Error("Panic recovered")

                    // エラーレスポンスを返す
                    c.JSON(500, map[string]string{
                        "error": "Internal Server Error",
                    })
                }
            }()

            // 5. ★重要★ ここで後続の全ての処理が実行される
            return next(c)
            // ↑ この中でhandler→usecase→repositoryが全て実行される
        }
    }
}
```

---

## 実行順序の詳細

### ケース1: 正常時の実行順序

```
1. Recovery Middleware 開始
   ├─ span作成
   ├─ defer recover()設定 ← 登録だけ（まだ実行されない）
   └─ next(c)呼び出し
      │
      2. CORS Middleware 開始
         ├─ span作成
         └─ next(c)呼び出し
            │
            3. Handler実行
               ├─ span作成
               └─ UseCase.CreateUser()呼び出し
                  │
                  4. UseCase実行
                     ├─ span作成
                     └─ Repository.Create()呼び出し
                        │
                        5. Repository実行
                           ├─ span作成
                           └─ DBにINSERT成功

                        6. Repository終了
                     7. UseCase終了
            8. Handler終了
         9. CORS Middleware終了
      10. Recovery Middleware終了
          └─ defer recover()実行（エラーなし）
```

### ケース2: Repository層でPanic発生時の実行順序

```
1. Recovery Middleware 開始
   ├─ span作成
   ├─ defer recover()設定 ← ★ここで設定（まだ実行されない）★
   └─ next(c)呼び出し
      │
      2. CORS Middleware 開始
         ├─ span作成
         └─ next(c)呼び出し
            │
            3. Handler実行
               ├─ span作成
               └─ UseCase.CreateUser()呼び出し
                  │
                  4. UseCase実行
                     ├─ span作成
                     └─ Repository.Create()呼び出し
                        │
                        5. Repository実行
                           ├─ span作成
                           └─ panic("database connection lost")

                        ← スタック巻き戻し（Repository終了）
                     ← スタック巻き戻し（UseCase終了）
            ← スタック巻き戻し（Handler終了）
         ← スタック巻き戻し（CORS Middleware終了）
      ← スタック巻き戻し（Recovery Middleware終了）

   ← ★defer recover()が実行されてPanicをキャッチ！★
   └─ エラーレスポンス送信 (HTTP 500)
```

### なぜキャッチできるのか

**答え**: `next(c)`の呼び出しによって、handler→usecase→repositoryの**すべての処理が同じゴルーチン・同じ関数スタックの中で実行される**ため、どこでPanicが発生しても、最初に設定した`defer recover()`がキャッチできます。

---

## Goの`defer`と`recover`の仕組み

### `defer`の特性

Goの`defer`は以下の特性を持ちます：

1. **関数が終了する瞬間に実行される**（正常終了でもPanicでも）
2. **複数の`defer`はLIFO（後入れ先出し）順で実行される**
3. **Panicが発生すると、スタックが巻き戻りながら各関数の`defer`が実行される**

### `recover`の特性

1. **`defer`の中でのみ有効**
2. **Panicをキャッチして通常のerrorとして扱える**
3. **Panicが発生した関数スタックを巻き戻す前にキャッチできる**

### コード例: deferとrecoverの基本

```go
func main() {
    fmt.Println("1. main開始")
    defer fmt.Println("6. main終了（defer）")

    functionA()

    fmt.Println("5. mainの残り処理")
}

func functionA() {
    fmt.Println("2. functionA開始")
    defer fmt.Println("4. functionA終了（defer）")

    functionB()

    // この行は実行されない（Panicのため）
    fmt.Println("この行は実行されない")
}

func functionB() {
    fmt.Println("3. functionB開始")
    panic("エラー発生！")

    // この行は実行されない
    fmt.Println("この行も実行されない")
}

// 出力:
// 1. main開始
// 2. functionA開始
// 3. functionB開始
// 4. functionA終了（defer）← スタック巻き戻り
// 6. main終了（defer）      ← スタック巻き戻り
// panic: エラー発生！
```

### recoverを追加した例

```go
func main() {
    fmt.Println("1. main開始")

    defer func() {
        if err := recover(); err != nil {
            fmt.Println("6. Panicをキャッチ:", err)
        }
    }()

    functionA()

    fmt.Println("7. mainの残り処理（Panic後も実行される）")
}

func functionA() {
    fmt.Println("2. functionA開始")
    defer fmt.Println("4. functionA終了（defer）")

    functionB()

    fmt.Println("この行は実行されない")
}

func functionB() {
    fmt.Println("3. functionB開始")
    panic("エラー発生！")
}

// 出力:
// 1. main開始
// 2. functionA開始
// 3. functionB開始
// 4. functionA終了（defer）      ← スタック巻き戻り
// 6. Panicをキャッチ: エラー発生！ ← recover()でキャッチ
// 7. mainの残り処理（Panic後も実行される）← 通常の処理に戻る
```

---

## ミドルウェアチェーンにおける`defer`と`recover`

### 実際のミドルウェアチェーンでの動作

```go
func RecoveryMiddleware(next HandlerFunc) HandlerFunc {
    return func(c Context) error {
        // ★ defer recover()を設定
        defer func() {
            if err := recover(); err != nil {
                fmt.Println("Panicをキャッチ:", err)
                c.JSON(500, map[string]string{"error": "Internal Server Error"})
            }
        }()

        // ★ 後続の全処理を実行
        return next(c)
        // ↑ この中でHandler→UseCase→Repositoryが全て実行される
        // ↑ どこでPanicが起きても、上のdefer recover()がキャッチする
    }
}

func Handler(c Context) error {
    fmt.Println("Handler実行")
    UseCase()
    return nil
}

func UseCase() {
    fmt.Println("UseCase実行")
    Repository()
}

func Repository() {
    fmt.Println("Repository実行")
    panic("データベース接続エラー") // ← ここでPanic
}

// 出力:
// Handler実行
// UseCase実行
// Repository実行
// Panicをキャッチ: データベース接続エラー ← recover()でキャッチ
// (500エラーレスポンス送信)
```

### 重要なポイント

1. **`next(c)`の中で実行される全ての処理は、同じゴルーチン内**
2. **同じ関数スタック内で実行される**
3. **Panicが発生すると、スタックが巻き戻り、最初に設定した`defer recover()`に到達する**
4. **`recover()`がPanicをキャッチし、エラーハンドリングを行う**

---

## ベストプラクティス

### 1. Recovery Middlewareは最も外側に配置

```go
// 正しい: Recoveryを最初に設定
e.Use(middleware.EchoRecoveryMiddleware())
e.Use(middleware.EchoCORSMiddleware())
e.Use(middleware.EchoLoggerMiddleware())

// 間違い: Recoveryが内側にある
e.Use(middleware.EchoLoggerMiddleware())
e.Use(middleware.EchoCORSMiddleware())
e.Use(middleware.EchoRecoveryMiddleware()) // ← これより前のMiddlewareでPanicが起きるとキャッチできない
```

### 2. Trace情報は事前に抽出

```go
func EchoRecoveryMiddleware() echo.MiddlewareFunc {
    return func(next echo.HandlerFunc) echo.HandlerFunc {
        return func(c echo.Context) error {
            span, ctx := tracer.StartSpanFromContext(c.Request().Context(), "middleware.recovery")
            defer span.Finish()

            // Panic発生前に情報を抽出
            var traceID, spanID string
            if span != nil {
                spanContext := span.Context()
                traceID = strconv.FormatUint(spanContext.TraceID(), 10)
                spanID = strconv.FormatUint(spanContext.SpanID(), 10)
            }

            defer func() {
                if err := recover(); err != nil {
                    // 抽出済みの情報を使用
                    logger.WithFields(logrus.Fields{
                        "dd.trace_id": traceID,
                        "dd.span_id":  spanID,
                    }).Error("Panic recovered")
                }
            }()

            return next(c)
        }
    }
}
```

### 3. スタックトレースを必ず記録

```go
defer func() {
    if err := recover(); err != nil {
        // スタックトレースを取得
        stackTrace := string(debug.Stack())

        logger.WithFields(logrus.Fields{
            "panic.value":       err,
            "panic.stack_trace": stackTrace, // ← これでエラー原因の特定が容易になる
        }).Error("Panic recovered")
    }
}()
```

### 4. エラーをDatadog Spanに記録

```go
defer func() {
    if err := recover(); err != nil {
        stackTrace := string(debug.Stack())

        // Spanにエラー情報を記録
        if span, ok := tracer.SpanFromContext(c.Request().Context()); ok {
            span.SetTag("error", true)
            span.SetTag("error.type", "panic")
            span.SetTag("error.msg", fmt.Sprintf("%v", err))
            span.SetTag("error.stack", stackTrace)
        }

        // ログ出力
        logger.Error("Panic recovered")
    }
}()
```

### 5. クライアントに適切なエラーレスポンスを返す

```go
defer func() {
    if err := recover(); err != nil {
        // ログ記録
        logger.Error("Panic recovered")

        // クライアントに500エラーを返す
        c.JSON(500, map[string]string{
            "error": "Internal Server Error",
            // 詳細なエラーメッセージは返さない（セキュリティリスク）
            // "detail": fmt.Sprintf("%v", err), // NG
        })
    }
}()
```

---

## まとめ

### Middleware チェーンの仕組み

1. **玉ねぎ構造**: 各Middlewareが次のMiddlewareをラップする
2. **同じゴルーチン・同じスタック**: `next(c)`の中で全処理が実行される
3. **スタック巻き戻り**: Panicが発生すると、最も外側の`defer recover()`に到達する

### なぜRecovery MiddlewareがすべてのPanicをキャッチできるのか

- **`defer func()`で`recover()`を設定** → 関数終了時（Panic時）に実行される
- **その後に`next(c)`を呼び出し** → 後続の全処理が同じスタック内で実行される
- **どこでPanicが起きても** → スタックが巻き戻り、最初の`defer recover()`に到達する

### 実装のポイント

1. Recovery Middlewareは**最も外側に配置**
2. Trace情報は**Panic発生前に抽出**
3. **スタックトレースを必ず記録**
4. エラー情報を**Datadog Spanに記録**
5. クライアントに**適切なエラーレスポンス**を返す

---

## 参考資料

- [本プロジェクトのRecovery Middleware実装](../../internal/presentation/middleware/echo_recovery.go)
- [Go公式ドキュメント: defer](https://go.dev/tour/flowcontrol/12)
- [Go公式ドキュメント: panic and recover](https://go.dev/blog/defer-panic-and-recover)
- [Echo Middleware Guide](https://echo.labstack.com/docs/middleware/)
