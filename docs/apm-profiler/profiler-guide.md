# Datadog Continuous Profiler完全ガイド

## 目次
1. [Profilerとは](#profilerとは)
2. [TracingとProfilerの違い](#tracingとprofilerの違い)
3. [実装方法（Go）](#実装方法go)
4. [Datadogでの確認方法](#datadogでの確認方法)
5. [パフォーマンス分析の実践](#パフォーマンス分析の実践)
6. [ベストプラクティス](#ベストプラクティス)

---

## Profilerとは

**Continuous Profiler（継続的プロファイリング）** は、アプリケーションのパフォーマンスをコードレベルで分析するツールです。

### Profilerができること

1. **CPU使用率の分析**
   - どの関数がCPU時間を消費しているか
   - ホットスポット（最も時間を使う箇所）の特定
   - 関数レベル、行レベルでの分析

2. **メモリ使用量の分析**
   - どの関数がメモリを大量に割り当てているか
   - メモリリークの検出
   - 不要なメモリ割り当ての発見

3. **Goroutine分析（Go特有）**
   - Goroutineの数と状態
   - Goroutineリークの検出

### なぜProfilerが必要か

**Tracingとの組み合わせで真価を発揮**

```
Tracing: "POST /api/users が500ms かかっている" ← 何が遅いかわかる
   ↓
Profiler: "user_repository.go:118 の rows.Scan() が300ms消費" ← なぜ遅いかわかる
   ↓
Profiler: "rows.Scan()で1リクエストあたり10MBのメモリ割り当て発生" ← 根本原因
```

---

## TracingとProfilerの違い

| 項目 | APM Tracing | Continuous Profiler |
|------|-------------|---------------------|
| **目的** | リクエストの流れを追跡 | コードレベルのパフォーマンス分析 |
| **粒度** | Handler → UseCase → Repository | 関数レベル・行レベル |
| **視点** | リクエスト単位 | 時間単位（CPU/メモリの使用状況） |
| **データ** | Span時間、エラー、タグ | CPU使用率、メモリ割り当て、Goroutine数 |
| **表示** | Flame Graph（リクエストの流れ） | Flame Graph（関数の実行時間） |
| **質問** | "どの処理が遅い？" | "どの関数が重い？なぜ遅い？" |
| **使い分け** | 処理フローの把握 | ボトルネックの特定と最適化 |

### 実践例

**シナリオ: APIが遅い**

1. **Tracingで調査**
   ```
   APM → Traces → POST /api/users
   → 平均レスポンスタイム: 500ms
   → mysql.create_user が 350ms かかっている
   ```

2. **Profilerで深掘り**
   ```
   Profiling → CPU → mysql.create_user を展開
   → user_repository.go:71 の db.QueryRowContext が 300ms
   → さらに展開すると rows.Scan() が遅い
   ```

3. **Profilerでメモリ確認**
   ```
   Profiling → Allocated Memory
   → rows.Scan() で10MB のメモリ割り当て発生
   → 不要なバッファリングが原因と判明
   ```

4. **最適化実施**
   - Scanの実装を見直し
   - バッファサイズを調整
   - 結果: 500ms → 150ms に改善

---

## 実装方法（Go）

### 1. 基本的な設定（main.go）

```go
package main

import (
    "os"
    "gopkg.in/DataDog/dd-trace-go.v1/profiler"
)

func main() {
    // Continuous Profilerの開始
    err := profiler.Start(
        profiler.WithService(os.Getenv("DD_SERVICE")),
        profiler.WithEnv(os.Getenv("DD_ENV")),
        profiler.WithVersion(os.Getenv("DD_VERSION")),
        profiler.WithProfileTypes(
            profiler.CPUProfile,      // CPU使用率プロファイル
            profiler.HeapProfile,      // メモリ割り当てプロファイル
        ),
    )
    if err != nil {
        logger.WithError(err).Warn("Failed to start profiler")
    }
    defer profiler.Stop()

    // アプリケーション起動
    startServer()
}
```

### 2. 利用可能なProfile Types

```go
profiler.WithProfileTypes(
    profiler.CPUProfile,           // CPU使用率（推奨）
    profiler.HeapProfile,          // メモリ割り当て（推奨）
    profiler.BlockProfile,         // ブロッキング操作
    profiler.MutexProfile,         // Mutex競合
    profiler.GoroutineProfile,     // Goroutine数と状態
)
```

**推奨設定**: `CPUProfile`と`HeapProfile`を有効化
- オーバーヘッドが低い（1-2%）
- 最も有用な情報が得られる

### 3. 環境変数での設定

```bash
# .env
DD_SERVICE=datadog-tour-api
DD_ENV=production
DD_VERSION=1.0.0
DD_PROFILING_ENABLED=true

# Profilerの詳細設定（オプション）
DD_PROFILING_CPU_DURATION=60s        # CPUプロファイルの収集間隔
DD_PROFILING_UPLOAD_PERIOD=60s       # Datadogへのアップロード間隔
```

### 4. Docker環境での設定

```yaml
# docker-compose.yml
services:
  api:
    environment:
      - DD_SERVICE=datadog-tour-api
      - DD_ENV=development
      - DD_VERSION=1.0.0
      - DD_PROFILING_ENABLED=true
      - DD_AGENT_HOST=datadog-agent
```

---

## Datadogでの確認方法

### 1. Profilingページへのアクセス

**方法A: 直接アクセス**
```
Datadog → APM → Profiling
```

**方法B: サービス経由**
```
Datadog → APM → Services → "datadog-tour-api" を選択 → Profiling タブ
```

### 2. CPU使用率の確認

**手順**:
1. **Profile Type**: `CPU` を選択
2. **Time Range**: 期間を設定（例：過去15分）
3. **Flame Graph** が表示される

**Flame Graphの読み方**:

```
┌───────────────────────────────────────────────────────────────┐
│ main.main [100%]                                              │ ← 一番下が起点
│ ┌─────────────────────────────────────────────────────────┐  │
│ │ http.Server.Serve [95%]                                 │  │
│ │ ┌─────────────────────────────────────────────────────┐ │  │
│ │ │ handler.CreateUser [60%] ← クリック可能              │ │  │
│ │ │ ┌───────────────────────────────────────────────┐   │ │  │
│ │ │ │ usecase.CreateUser [55%]                      │   │ │  │
│ │ │ │ ┌─────────────────────────┐ ┌───────────┐    │   │ │  │
│ │ │ │ │ mysql.Create [40%] ← 重い│ │redis [10%]│    │   │ │  │
│ │ │ │ │ ┌─────────────────────┐ │ └───────────┘    │   │ │  │
│ │ │ │ │ │ rows.Scan [35%] ← ホットスポット         │   │ │  │
│ │ │ │ │ └─────────────────────┘ │                  │   │ │  │
│ │ │ │ └─────────────────────────┘                  │   │ │  │
│ │ │ └───────────────────────────────────────────────┘   │ │  │
│ │ └─────────────────────────────────────────────────────┘ │  │
│ └─────────────────────────────────────────────────────────┘  │
└───────────────────────────────────────────────────────────────┘

横幅 = CPU時間の割合（広い = 時間を多く消費）
縦   = 関数の呼び出しスタック（下が呼び出し元、上が呼び出し先）
色   = パッケージごとの色分け
```

**Spanをクリックすると表示される情報**:
- **Self CPU**: その関数自体が消費したCPU時間（子関数を除く）
- **Total CPU**: 子関数を含めた合計CPU時間
- **Samples**: サンプリング回数
- **File**: ファイル名と行番号
- **Function**: 関数名

### 3. メモリ使用量の確認

**手順**:
1. **Profile Type**: `Allocated Memory` を選択
2. **Flame Graph** で確認

**Flame Graphの見方**（CPUと同じ構造）:
```
横幅 = メモリ割り当て量の割合
縦   = 関数の呼び出しスタック

例：
rows.Scan の横幅が広い
→ この関数で大量のメモリを割り当てている
→ クリックして詳細を確認
→ 1回の呼び出しで10MBを割り当てていることが判明
```

**確認すべき指標**:
- **Self Allocated**: その関数自体が割り当てたメモリ量
- **Total Allocated**: 子関数を含めた合計メモリ量
- **Objects**: 割り当てたオブジェクト数

### 4. Goroutine数の確認（Go特有）

**手順**:
1. **Profile Type**: `Goroutines` を選択
2. グラフでGoroutine数の推移を確認

**確認ポイント**:
```
正常: Goroutine数が一定または緩やかな増減
異常: Goroutine数が右肩上がりに増加 → Goroutineリーク
```

### 5. コードビューでの詳細確認

**手順**:
1. Flame Graphで関数をクリック
2. 右パネルに詳細情報が表示される
3. **View Code** ボタンをクリック（利用可能な場合）
4. 行単位のCPU/メモリ使用率が表示される

**例**:
```go
// user_repository.go
func (r *UserRepository) FindAll(ctx context.Context) ([]*entities.User, error) {
    query := "SELECT * FROM users"
    rows, err := r.db.QueryContext(ctx, query)  // CPU: 5%
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var users []*entities.User
    for rows.Next() {                            // CPU: 10%
        var user entities.User
        if err := rows.Scan(&user.ID, &user.Name); err != nil { // CPU: 70% ← ホットスポット！
            continue
        }
        users = append(users, &user)             // Memory: 15MB
    }
    return users, nil
}
```

---

## パフォーマンス分析の実践

### 実践例1: 遅いエンドポイントの最適化

**ステップ1: Tracingで問題を発見**
```
APM → Traces → POST /api/users
→ 平均レスポンスタイム: 500ms
→ mysql.create_user スパンが 350ms
```

**ステップ2: Profilingで関数を特定**
```
Profiling → CPU → 期間を選択
→ Flame Graphで mysql.create_user を探す
→ クリックして展開
→ db.QueryRowContext が全体の60%を占めている
```

**ステップ3: 行レベルで確認**
```
→ View Code をクリック
→ user_repository.go:71 の Scanが遅いと判明
```

**ステップ4: メモリも確認**
```
Profiling → Allocated Memory
→ 同じ関数で10MBのメモリ割り当て
→ 不要なバッファリングが原因
```

**ステップ5: 最適化実施**
```go
// Before: 不要な中間バッファ
var buffer []byte
rows.Scan(&buffer)
user.Name = string(buffer)  // ← ここでコピーが発生

// After: 直接Scan
rows.Scan(&user.Name)  // ← コピー不要
```

**結果**:
- CPU時間: 350ms → 50ms（86%削減）
- メモリ: 10MB → 1MB（90%削減）

### 実践例2: メモリリークの検出

**症状**: サーバーのメモリ使用量が徐々に増加

**ステップ1: Profilingでメモリを確認**
```
Profiling → Heap → 24時間の推移を確認
→ メモリが右肩上がりに増加
```

**ステップ2: Flame Graphで原因を特定**
```
→ cacheManager.Set が大量のメモリを保持
→ クリックして詳細を確認
→ キャッシュがクリアされていない
```

**ステップ3: コードを確認**
```go
// Before: TTLがない
func (c *CacheManager) Set(key string, value interface{}) {
    c.cache[key] = value  // ← クリアされない
}

// After: TTL追加
func (c *CacheManager) Set(key string, value interface{}, ttl time.Duration) {
    c.cache[key] = cacheEntry{
        value:     value,
        expiresAt: time.Now().Add(ttl),
    }
}

// 定期的にクリア
func (c *CacheManager) cleanup() {
    ticker := time.NewTicker(1 * time.Minute)
    for range ticker.C {
        c.removeExpired()
    }
}
```

**結果**: メモリリークが解消

### 実践例3: N+1クエリ問題の発見

**ステップ1: Tracingで同じクエリを発見**
```
APM → Traces → GET /api/orders
→ mysql.get_user スパンが100回実行されている
```

**ステップ2: Profilingで影響を確認**
```
Profiling → CPU
→ db.QueryRow が全体の80%を占める
→ 1回1回は速いが、合計で800msかかっている
```

**ステップ3: 最適化実施**
```go
// Before: N+1クエリ
func GetOrders(ctx context.Context) ([]*Order, error) {
    orders, _ := getOrders(ctx)
    for _, order := range orders {
        user, _ := getUserByID(ctx, order.UserID)  // ← 100回実行
        order.User = user
    }
    return orders, nil
}

// After: JOIN または IN句
func GetOrders(ctx context.Context) ([]*Order, error) {
    query := `
        SELECT o.*, u.name, u.email
        FROM orders o
        JOIN users u ON o.user_id = u.id
    `
    // 1回のクエリで全て取得
}
```

**結果**: 800ms → 50ms（93%削減）

---

## ベストプラクティス

### 1. 本番環境で常時有効化

```go
// 推奨: 本番環境でも有効化
// オーバーヘッドは1-2%程度で無視できる
err := profiler.Start(
    profiler.WithService(os.Getenv("DD_SERVICE")),
    profiler.WithEnv(os.Getenv("DD_ENV")),
    profiler.WithProfileTypes(
        profiler.CPUProfile,
        profiler.HeapProfile,
    ),
)

// 非推奨: 本番環境で無効化
if os.Getenv("ENV") != "production" {
    profiler.Start(...)  // 本番の問題を検出できない
}
```

### 2. 適切なProfile Typeを選択

```go
// 推奨: CPUとHeapのみ（オーバーヘッド低い）
profiler.WithProfileTypes(
    profiler.CPUProfile,
    profiler.HeapProfile,
)

// 注意: 全て有効化するとオーバーヘッド増加
profiler.WithProfileTypes(
    profiler.CPUProfile,
    profiler.HeapProfile,
    profiler.BlockProfile,     // 必要な場合のみ
    profiler.MutexProfile,     // 必要な場合のみ
    profiler.GoroutineProfile, // 必要な場合のみ
)
```

### 3. Tracingと組み合わせる

```
Tracingで「何が」遅いかを特定
   ↓
Profilerで「なぜ」遅いかを分析
   ↓
最適化実施
   ↓
Tracingで効果を確認
```

### 4. 定期的な確認

```
週次: 主要エンドポイントのCPU/メモリプロファイルを確認
月次: 全体的なパフォーマンストレンドを分析
デプロイ後: 新しいコードのパフォーマンスを確認
```

### 5. アラート設定

```
Datadog Monitors で設定:
- メモリ使用率が70%を超えた場合
- CPU使用率が80%を超えた場合
- Goroutine数が1000を超えた場合
```

---

## よくある質問

### Q1: Profilerのオーバーヘッドはどのくらい？

**A**: 約1-2%程度
- CPUProfile: 1-2%
- HeapProfile: 1%未満
- 本番環境でも常時有効化して問題ない

### Q2: データはいつDatadogに表示される？

**A**: 数分後
- プロファイルは60秒ごとに収集
- Datadogへのアップロードに数分かかる
- リアルタイムではない

### Q3: ローカル環境でProfilerは動く？

**A**: 動く
- Datadog Agentが起動していれば動作
- docker-compose で datadog-agent を起動

### Q4: Profilerとpprofの違いは？

**A**:
| 項目 | Datadog Profiler | pprof |
|------|------------------|-------|
| 常時有効化 | ○（オーバーヘッド低い） | × |
| 本番環境 | ○ | △（手動で有効化） |
| 可視化 | ○（Flame Graph） | △（手動でツール起動） |
| 履歴 | ○（時系列で確認可能） | × |
| 統合 | ○（APMと統合） | × |

### Q5: Flame Graphが表示されない場合は？

**確認ポイント**:
1. profiler.Start()が呼ばれているか
2. Datadog Agentが起動しているか
3. 環境変数が設定されているか
4. 十分な負荷がかかっているか（サンプリングベース）
5. 数分待ったか（リアルタイムではない）

---

## まとめ

### Profilerの役割
- **コードレベルのパフォーマンス分析**
- **CPU使用率とメモリ割り当ての可視化**
- **ホットスポット（ボトルネック）の特定**

### 実装のポイント
1. `profiler.Start()`で初期化
2. `CPUProfile`と`HeapProfile`を有効化
3. 本番環境でも常時有効化（オーバーヘッド1-2%）
4. Tracingと組み合わせて使用

### Datadogでの確認
1. APM → Profiling → サービス選択
2. Profile Type（CPU / Allocated Memory）を選択
3. Flame Graphで関数ごとの使用率を確認
4. 関数をクリックして詳細情報を確認

### パフォーマンス分析の流れ
```
Tracing: 何が遅い？ → POST /api/users が500ms
   ↓
Profiler: どの関数が重い？ → rows.Scan()が300ms
   ↓
Profiler: なぜ重い？ → 10MBのメモリ割り当て
   ↓
最適化: バッファリング削減
   ↓
Tracing: 効果確認 → 500ms → 150ms
```

---

## 参考資料

- [Datadog Profiler Documentation](https://docs.datadoghq.com/profiler/)
- [dd-trace-go Profiler](https://github.com/DataDog/dd-trace-go/tree/v1/profiler)
- [本プロジェクトの実装例](../cmd/api/main.go)
