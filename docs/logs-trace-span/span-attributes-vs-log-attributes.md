# Span Attributes vs Log Attributes

Datadogでは、**APM Traces（Span Attributes）** と **Logs（Event Attributes）** という2つの異なるデータストアがあり、それぞれ異なる目的で使用されます。

## 目次

- [1. データフローの違い](#1-データフローの違い)
- [2. APM Traces - Span Attributes](#2-apm-traces---span-attributes)
- [3. Logs Explorer - Event Attributes](#3-logs-explorer---event-attributes)
- [4. Log-Trace Correlation](#4-log-trace-correlation)
- [5. 実際のデータの流れ](#5-実際のデータの流れ)
- [6. 使い分けのベストプラクティス](#6-使い分けのベストプラクティス)

---

## 1. データフローの違い

### APM Traces (Span Attributes)

```
[Go Application]
    ↓ span.SetTag("user.id", 123)
[dd-trace-go library]
    ↓ span.Finish() 時に送信
[Datadog Agent :8126] (APM port)
    ↓
[Datadog APM Backend]
    ↓
[APM > Traces > Span 詳細]
```

### Logs (Event Attributes)

```
[Go Application]
    ↓ logger.InfoContext(..., "user.id", 123) → JSON出力
[標準出力 (stdout)]
    ↓ {"user.id": 123, "msg": "..."}
[Docker Container]
    ↓
[Datadog Agent :10518] (Logs port)
    ↓ JSON自動パース
[Datadog Logs Backend]
    ↓
[Logs > Log Explorer]
```

---

## 2. APM Traces - Span Attributes

### コード例

```go
span, ctx := tracer.StartSpanFromContext(ctx, "usecase.get_user")
defer span.Finish()

// Span Attributesを設定
span.SetTag("user.id", 123)
span.SetTag("cache.hit", true)
span.SetTag("data.source", "cache")
```

### 送信されるデータ

**送信先:** Datadog Agent `:8126` (APM port)

**データ形式:**
```
Span: usecase.get_user
Duration: 15.2ms
Tags:
  - user.id: 123
  - cache.hit: true
  - data.source: cache
  - dd.trace_id: 1234567890123456789
  - dd.span_id: 9876543210987654321
```

### 表示場所

- **APM > Traces** でトレースを検索
- Flame Graph上でSpanをクリック
- Span詳細の `Tags` セクションに表示

### 用途

| 用途 | 説明 | 例 |
|------|------|-----|
| **パフォーマンス分析** | レイテンシ、エラー率の集計 | P95レイテンシが高いエンドポイントを特定 |
| **トレース検索** | 特定条件でトレースをフィルタリング | `user.id:123` でユーザー別トレース抽出 |
| **メトリクス化** | カスタムメトリクスの生成 | キャッシュヒット率を時系列グラフ化 |
| **Service Map** | サービス間の依存関係可視化 | どのサービスがどこを呼んでいるか |

### 保存期間

- デフォルト: 15日間（プランによる）
- サンプリングされる可能性あり（全トレースは保存されない場合も）
- Retention Filters で保存期間をカスタマイズ可能

### 設定方法（このプロジェクトでの実装）

```go
// internal/usecase/user_usecase.go

// ビジネスメトリクスとして有用
span.SetTag("user.id", user.ID)         // ユーザー識別
span.SetTag("cache.hit", true)          // キャッシュヒット率
span.SetTag("cache.set", true)          // キャッシュ設定成功率
span.SetTag("data.source", "database")  // データソース追跡
span.SetTag("users.count", 100)         // クエリ結果数

// 自動設定されるため不要（logging.LogErrorWithTrace内で設定）
// span.SetTag("error", true)
// span.SetTag("error.msg", err.Error())
```

---

## 3. Logs Explorer - Event Attributes

### コード例

```go
logging.LogWithTrace(ctx, logger, "usecase", "User found in cache", map[string]any{
    "user.id":   user.ID,
    "user.name": user.Name,
    "cache.key": cacheKey,
})
```

### 送信されるデータ

**送信先:** Datadog Agent `:10518` (Logs port)

**データ形式（JSON）:**
```json
{
  "time": "2024-11-06T15:04:05Z",
  "level": "INFO",
  "msg": "[usecase] User found in cache",
  "file": "/path/to/user_usecase.go",
  "line": 101,
  "dd.trace_id": "1234567890123456789",
  "dd.span_id": "9876543210987654321",
  "layer": "usecase",
  "user.id": 123,
  "user.name": "John Doe",
  "cache.key": "user:123"
}
```

### 表示場所

- **Logs > Log Explorer** で検索
- 各ログイベントの詳細ビュー
- `Event Attributes` セクションに全フィールドが表示
- サイドバーでファセット検索可能

### 用途

| 用途 | 説明 | 例 |
|------|------|-----|
| **デバッグ** | 詳細なコンテキスト情報 | エラー発生時の変数値を確認 |
| **監査** | 誰が何をしたかの記録 | ユーザーの操作履歴を追跡 |
| **フルテキスト検索** | エラーメッセージ内を検索 | 特定のエラーパターンを検索 |
| **アラート** | 特定ログパターンでアラート | `error.notify:true` でアラート |

### 保存期間

- デフォルト: 15日間（プランによる）
- **サンプリングなし**（ただし、フィルタで除外は可能）
- Index設定で保存期間をカスタマイズ可能

### 設定方法（このプロジェクトでの実装）

```go
// internal/common/logging/logger.go

// 通常のログ
logging.LogWithTrace(ctx, logger, "usecase", "Creating user", map[string]any{
    "user.name":  name,
    "user.email": email,
})

// エラーログ（アラート対象）
logging.LogErrorWithTrace(ctx, logger, "usecase", "Failed to create user", err, map[string]any{
    "user.id": userID,
})
// ↑ 内部で自動的に span.SetTag("error", true) も設定

// エラーログ（アラート対象外）
logging.LogErrorWithTraceNotNotify(ctx, logger, "usecase", "User already exists", err, map[string]any{
    "error.type": "validation_error",
})
```

---

## 4. Log-Trace Correlation

### 相関の仕組み

APM TracesとLogsを繋ぐのが **`dd.trace_id`** と **`dd.span_id`** です。

```go
// internal/common/logging/logger.go:26-31
if span, ok := tracer.SpanFromContext(ctx); ok {
    spanContext := span.Context()
    attrs = append(attrs,
        "dd.trace_id", strconv.FormatUint(spanContext.TraceID(), 10),  // ← ログとトレースを紐付け
        "dd.span_id", strconv.FormatUint(spanContext.SpanID(), 10),
    )
}
```

### 相関により実現できること

#### 1. Logs → Traces

Logs Explorerでログを開くと、右側に **"View Trace"** ボタンが表示されます:

```
[Logs Explorer]
  Log: "[usecase] User found in cache"
  Attributes:
    - user.id: 123
    - dd.trace_id: 1234567890...  ← これをクリック

  → APM Traces に自動遷移
```

#### 2. Traces → Logs

APM Tracesでトレースを開くと、**"View Logs"** ボタンが表示されます:

```
[APM Traces]
  Span: usecase.get_user
  Tags:
    - user.id: 123
    - dd.trace_id: 1234567890...  ← これをクリック

  → Logs Explorer に自動遷移（このトレースIDのログのみ表示）
```

### Datadogでの確認方法

1. **Logs Explorerでログを検索**
   ```
   @user.id:123
   ```

2. **ログをクリックして詳細を開く**

3. **右上の "View Trace" ボタンをクリック**
   - そのログが発生した時の完全なトレースが表示される
   - どの処理で時間がかかっているかが一目瞭然

4. **トレースから戻る場合は "View Logs" をクリック**
   - そのトレース中に出力された全ログが表示される

---

## 5. 実際のデータの流れ

### アプリケーションコード

```go
// internal/usecase/user_usecase.go
func (uc *UserUseCase) GetUser(ctx context.Context, id int) (*entities.User, error) {
    span, ctx := tracer.StartSpanFromContext(ctx, "usecase.get_user")
    defer span.Finish()

    // ❶ Span Attributesを設定（APMに送信）
    span.SetTag("user.id", id)
    span.SetTag("cache.hit", true)
    span.SetTag("data.source", "cache")

    // ❷ Log Attributesを設定（Logsに送信）
    logging.LogWithTrace(ctx, logger, "usecase", "User found in cache", map[string]any{
        "user.id":   id,
        "user.name": "John Doe",
        "cache.key": "user:123",
    })

    return user, nil
}
```

### 結果

#### APM Traces - Span詳細

```
Span: usecase.get_user
Duration: 2.5ms
Status: OK

Tags:
  - user.id: 123          ← span.SetTag()で設定
  - cache.hit: true       ← span.SetTag()で設定
  - data.source: cache    ← span.SetTag()で設定
  - dd.trace_id: 1234567890123456789
  - dd.span_id: 9876543210987654321
```

#### Logs Explorer - Event Attributes

```json
{
  "time": "2024-11-06T15:04:05.123Z",
  "level": "INFO",
  "msg": "[usecase] User found in cache",
  "file": "/app/internal/usecase/user_usecase.go",
  "line": 101,
  "dd.trace_id": "1234567890123456789",  ← トレースと紐付け
  "dd.span_id": "9876543210987654321",
  "layer": "usecase",
  "user.id": 123,           ← logging.LogWithTrace()で設定
  "user.name": "John Doe",  ← logging.LogWithTrace()で設定
  "cache.key": "user:123"   ← logging.LogWithTrace()で設定
}
```

### 両方に含まれるもの

| フィールド | Span Attributes | Log Attributes | 目的 |
|-----------|----------------|---------------|------|
| `user.id` | Yes | Yes | 両方で検索可能にする |
| `dd.trace_id` | Yes (自動) | Yes (自動) | 相関のため |
| `dd.span_id` | Yes (自動) | Yes (自動) | 相関のため |

### Spanのみに含まれるもの

| フィールド | 理由 |
|-----------|------|
| `cache.hit` | メトリクス化してキャッシュヒット率を監視 |
| `data.source` | パフォーマンス分析（DB vs Cache） |

### Logのみに含まれるもの

| フィールド | 理由 |
|-----------|------|
| `user.name` | デバッグ用の詳細情報（PII） |
| `cache.key` | デバッグ用の詳細情報 |
| `file`, `line` | エラー発生箇所の特定 |

---

## 6. 使い分けのベストプラクティス

### 基本原則

| 情報の種類 | Span Attributes | Log Attributes | 理由 |
|-----------|----------------|---------------|------|
| **ユーザーID** | 推奨 | 推奨 | 検索・集計・デバッグの両方で必要 |
| **キャッシュヒット** | 推奨 | オプション | メトリクス化して監視したい |
| **データソース** | 推奨 | オプション | パフォーマンス分析に使用 |
| **ユーザー名** | 不要 | 推奨 | デバッグ用。PIIなのでSpanには含めない |
| **メールアドレス** | 不要 | 推奨 | デバッグ用。PIIなのでSpanには含めない |
| **エラーメッセージ** | 不要 | 推奨 | logging関数が自動設定 |
| **スタックトレース** | 不要 | 推奨 | デバッグ用。サイズが大きい |
| **SQLクエリ** | 不要 | 推奨 | デバッグ用。logging関数が自動設定 |
| **リクエストボディ** | 不要 | 注意 | デバッグ用だが、サイズとPIIに注意 |

### Span Attributesに含めるべきもの

**軽量なビジネスメトリクス**
```go
span.SetTag("user.id", 123)           // ID（識別子）
span.SetTag("cache.hit", true)        // boolean
span.SetTag("data.source", "cache")   // enum的な値
span.SetTag("query.rows_count", 100)  // 数値
```

**含めるべきでないもの**
```go
span.SetTag("user.name", "John Doe")          // PII（個人情報）
span.SetTag("user.email", "john@example.com") // PII
span.SetTag("error.msg", err.Error())         // logging関数が自動設定
span.SetTag("stack.trace", stackTrace)        // サイズが大きい
span.SetTag("request.body", jsonString)       // サイズが大きい
```

### Log Attributesに含めるべきもの

**デバッグに必要な詳細情報**
```go
logging.LogWithTrace(ctx, logger, "usecase", "Creating user", map[string]any{
    "user.id":    123,                    // 識別子
    "user.name":  "John Doe",             // デバッグ用
    "user.email": "john@example.com",     // デバッグ用
    "cache.key":  "user:123",             // デバッグ用
    "retry.count": 3,                     // デバッグ用
})
```

**含めるべきでないもの**
```go
map[string]any{
    "user.password": "secret123",         // 絶対ダメ！
    "api.token": "Bearer abc123...",      // 絶対ダメ！
    "credit.card": "1234-5678-9012-3456", // 絶対ダメ！
}
```

### このプロジェクトでの実装例

#### 良い例

```go
// internal/usecase/user_usecase.go

// Span: ビジネスメトリクスのみ
span.SetTag("user.id", user.ID)
span.SetTag("cache.hit", true)
span.SetTag("cache.set", true)
span.SetTag("data.source", "database")

// Log: 詳細なデバッグ情報
logging.LogWithTrace(ctx, logger, "usecase", "User found in cache", map[string]any{
    "user.id":   user.ID,
    "cache.key": cacheKey,
})

// エラー時: logging関数が自動的にspanにもerror情報を設定
logging.LogErrorWithTrace(ctx, logger, "usecase", "Failed to get user", err, map[string]any{
    "user.id": id,
})
// ↑ 内部で自動的に以下が実行される:
//   span.SetTag("error", true)
//   span.SetTag("error.msg", err.Error())
//   span.SetTag("error.notify", true)
```

#### 悪い例（修正前）

```go
// 重複: logging関数が自動設定するのに手動で設定
span.SetTag("error", true)              // ← 不要
span.SetTag("error.msg", err.Error())   // ← 不要
logging.LogErrorWithTrace(...)

// PII: Spanに個人情報を含める
span.SetTag("user.name", user.Name)     // ← PIIなので削除
span.SetTag("user.email", user.Email)   // ← PIIなので削除
```

---

## まとめ

### Span Attributes (`span.SetTag()`)

- **目的**: パフォーマンス監視、メトリクス化
- **送信先**: Datadog APM Backend
- **表示**: APM > Traces
- **データサイズ**: 軽量に保つ（ID、boolean、enum、数値）
- **PII**: 含めない（user.id のみOK）

### Log Attributes (JSON log)

- **目的**: デバッグ、監査、エラー調査
- **送信先**: Datadog Logs Backend
- **表示**: Logs > Log Explorer
- **データサイズ**: 詳細情報OK（ただし機密情報は除く）
- **PII**: 必要に応じて含める（user.name, user.email など）

### Log-Trace Correlation

- **キー**: `dd.trace_id`, `dd.span_id`
- **効果**: ログとトレースを相互に行き来できる
- **実装**: logging関数が自動的に設定

### 両方使うのがベストプラクティス

- **Span**: 軽量なビジネスメトリクス
- **Log**: 詳細なデバッグ情報
- **相関**: `dd.trace_id` で繋ぐ

これにより、パフォーマンス分析とデバッグの両方を効率的に行えます。
