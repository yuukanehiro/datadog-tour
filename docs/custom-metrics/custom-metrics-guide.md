# Datadog Custom Metrics完全ガイド

## 目次
1. [Custom Metricsとは](#custom-metricsとは)
2. [DogStatsDの仕組み](#dogstatsdの仕組み)
3. [メトリクスの種類](#メトリクスの種類)
4. [実装方法](#実装方法)
5. [このプロジェクトで実装されているメトリクス](#このプロジェクトで実装されているメトリクス)
6. [Datadogでの確認方法](#datadogでの確認方法)
7. [ダッシュボード作成](#ダッシュボード作成)
8. [ベストプラクティス](#ベストプラクティス)

---

## Custom Metricsとは

Custom Metricsは、アプリケーション固有のビジネスメトリクスやパフォーマンス指標を収集・可視化する機能です。

### なぜCustom Metricsが必要か

Datadogは標準でインフラメトリクス（CPU、メモリ、ネットワーク等）を収集しますが、**ビジネスロジックやアプリケーション固有の指標**は自分で実装する必要があります。

**例：ECサイトの場合**
- 注文数（Counter）
- カート内の商品数（Gauge）
- 決済処理時間（Timing）
- アクティブユーザー数（Gauge）
- 在庫切れ商品数（Gauge）

### APM TracesとCustom Metricsの違い

| 項目 | APM Traces | Custom Metrics |
|------|-----------|----------------|
| **目的** | リクエストの流れを追跡 | 定量的な指標を収集 |
| **データ** | 個々のリクエスト | 集計された数値 |
| **保存期間** | 15日間（デフォルト） | 15ヶ月 |
| **用途** | エラー調査、パフォーマンス分析 | トレンド分析、アラート |
| **例** | 特定のリクエストがどこで遅いか | 1時間あたりの平均リクエスト数 |

---

## DogStatsDの仕組み

### アーキテクチャ

```
┌─────────────────────┐
│  Go Application     │
│  ┌───────────────┐  │
│  │ statsd.Incr() │  │  UDP (Port 8125)
│  └───────┬───────┘  │  ───────────────┐
└──────────┼──────────┘                  │
           │                             ▼
           │                  ┌──────────────────┐
           │                  │  Datadog Agent   │
           │                  │  (DogStatsD)     │
           │                  └────────┬─────────┘
           │                           │
           │                           │ HTTPS
           │                           │
           │                           ▼
           │                  ┌──────────────────┐
           │                  │  Datadog API     │
           │                  │  (クラウド)       │
           │                  └──────────────────┘
```

### DogStatsDとは

- **StatsD Protocol**: Etsyが開発したメトリクス収集プロトコル
- **DogStatsD**: DatadogによるStatsD実装（拡張版）
- **UDP通信**: 軽量・非同期でアプリケーションのパフォーマンスに影響しない
- **Aggregation**: Datadog Agentがメトリクスを集計してDatadogに送信

### 通信フロー

1. **アプリケーション**: `statsdClient.Incr("api.users.create", nil, 1)`を呼び出し
2. **UDP送信**: メトリクスをUDP（Port 8125）でDatadog Agentに送信
3. **Agent集計**: Datadog Agentが10秒間隔でメトリクスを集計
4. **クラウド送信**: 集計されたメトリクスをDatadog APIに送信（HTTPS）
5. **可視化**: Datadog UIでメトリクスを確認

---

## メトリクスの種類

DogStatsDは5種類のメトリクスタイプをサポートしています。

### 1. Counter（カウンター）

**用途**: イベントの発生回数をカウント

**特徴**:
- 増加のみ（デクリメントも可能だが非推奨）
- Datadog Agent側で**レート（rate）に変換**される（例: 1秒あたりの発生回数）

**実装例**:
```go
// ユーザー作成回数をカウント
statsdClient.Incr("api.users.create", nil, 1)

// タグ付きでカウント
statsdClient.Incr("api.requests", []string{"endpoint:/api/users", "method:POST"}, 1)
```

**Datadogでの表示**:
- `api.users.create.count`: 累積値
- `api.users.create.rate`: 1秒あたりの発生回数

### 2. Gauge（ゲージ）

**用途**: 現在の値をそのまま記録（スナップショット）

**特徴**:
- 最新の値で上書きされる
- 増加・減少両方可能

**実装例**:
```go
// 現在のユーザー数
statsdClient.Gauge("api.users.total", float64(userCount), nil, 1)

// メモリ使用量
statsdClient.Gauge("app.memory.used", float64(memStats.Alloc), nil, 1)

// アクティブ接続数
statsdClient.Gauge("api.connections.active", float64(activeConns), nil, 1)
```

**Datadogでの表示**:
- 最新の値がそのまま表示される

### 3. Timing（タイミング）

**用途**: 処理時間の計測

**特徴**:
- 自動的に統計値（平均、最小、最大、パーセンタイル）を計算
- `time.Duration`をそのまま渡せる

**実装例**:
```go
// 処理時間を計測
start := time.Now()
users, err := uc.RUser.FindAll(ctx)
duration := time.Since(start)

// Timingメトリクスを送信
statsdClient.Timing("api.users.list.duration", duration, nil, 1)
```

**Datadogでの表示**:
- `api.users.list.duration.avg`: 平均処理時間
- `api.users.list.duration.p95`: 95パーセンタイル
- `api.users.list.duration.max`: 最大処理時間

### 4. Histogram（ヒストグラム）

**用途**: 値の分布を記録（Timingの汎用版）

**特徴**:
- 複数の統計値を自動計算（平均、合計、最小、最大、パーセンタイル等）
- タイミング以外の値にも使用可能

**実装例**:
```go
// リクエストサイズの分布
statsdClient.Histogram("api.request.size", float64(requestSize), nil, 1)

// レスポンスサイズの分布
statsdClient.Histogram("api.response.size", float64(len(responseBody)), nil, 1)
```

### 5. Set（セット）

**用途**: ユニークな値の数をカウント

**特徴**:
- 重複を自動的に排除
- ユニークユーザー数、IPアドレス数等に使用

**実装例**:
```go
// ユニークなユーザーIDをカウント
statsdClient.Set("api.users.unique", fmt.Sprintf("%d", userID), nil, 1)
```

### メトリクスタイプの選び方

| 計測したい内容 | メトリクスタイプ | 例 |
|-------------|----------------|-----|
| イベント発生回数 | **Counter** | ユーザー作成数、エラー数 |
| 現在の状態・値 | **Gauge** | アクティブユーザー数、在庫数 |
| 処理時間 | **Timing** | API応答時間、DB処理時間 |
| 値の分布 | **Histogram** | リクエストサイズ、レスポンスサイズ |
| ユニーク数 | **Set** | ユニークユーザー数、ユニークIP数 |

---

## 実装方法

### 1. DogStatsD Clientの初期化

**場所**: `cmd/api/main.go`

```go
package main

import (
    "fmt"
    "os"
    "github.com/DataDog/datadog-go/v5/statsd"
)

func main() {
    // Initialize DogStatsD client
    statsdClient, err := statsd.New(
        fmt.Sprintf("%s:%s", os.Getenv("DD_AGENT_HOST"), "8125"),
        statsd.WithTags([]string{
            "env:" + os.Getenv("DD_ENV"),
            "service:" + os.Getenv("DD_SERVICE"),
        }),
    )
    if err != nil {
        logger.WithError(err).Fatal("Failed to initialize StatsD client")
    }
    defer statsdClient.Close()

    // ... アプリケーション起動
}
```

**重要なポイント**:
- **ポート 8125**: DogStatsDのデフォルトポート（UDP）
- **グローバルタグ**: すべてのメトリクスに自動的に付与される
- **defer Close()**: アプリケーション終了時にクライアントをクローズ

### 2. RepoLocatorへの追加

**場所**: `internal/common/context/context.go`

```go
package context

import (
    "github.com/DataDog/datadog-go/v5/statsd"
)

// RepoLocator holds all repositories and shared dependencies
type RepoLocator struct {
    UserRepo     port.UserRepository
    CacheRepo    port.CacheRepository
    StatsdClient *statsd.Client  // 追加
}
```

### 3. Handlerでの使用

**場所**: `internal/presentation/interface-adapter/handler/user_handler.go`

```go
func (h *UserHandler) CreateUser(c echo.Context) error {
    span, ctx := tracer.StartSpanFromContext(c.Request().Context(), "handler.create_user")
    defer span.Finish()

    logger := appcontext.GetLogger(ctx)
    repoLocator := appcontext.GetRepoLocator(ctx)

    // ユーザー作成試行をカウント
    if repoLocator != nil && repoLocator.StatsdClient != nil {
        repoLocator.StatsdClient.Incr("api.users.create", nil, 1)
    }

    // ビジネスロジック実行
    user, err := interactor.CreateUser(ctx, req.Name, req.Email)
    if err != nil {
        // エラーをカウント
        if repoLocator != nil && repoLocator.StatsdClient != nil {
            repoLocator.StatsdClient.Incr("api.users.create.error", nil, 1)
        }
        return c.JSON(http.StatusInternalServerError, errorResponse)
    }

    // 成功をカウント
    if repoLocator != nil && repoLocator.StatsdClient != nil {
        repoLocator.StatsdClient.Incr("api.users.create.success", nil, 1)
    }

    return c.JSON(http.StatusCreated, user)
}
```

### 4. UseCaseでの使用

**場所**: `internal/usecase/user_usecase.go`

```go
func (uc *UserUseCase) GetUser(ctx context.Context, id int) (*entities.User, error) {
    span, ctx := tracer.StartSpanFromContext(ctx, "usecase.get_user")
    defer span.Finish()

    // キャッシュ確認
    cacheKey := fmt.Sprintf("user:%d", id)
    cachedData, err := uc.RCache.Get(ctx, cacheKey)

    if err == nil && cachedData != "" {
        // キャッシュヒット
        if repoLocator := appcontext.GetRepoLocator(ctx); repoLocator != nil && repoLocator.StatsdClient != nil {
            repoLocator.StatsdClient.Incr("api.users.get.cache_hit", nil, 1)
        }
        return &user, nil
    }

    // キャッシュミス
    if repoLocator := appcontext.GetRepoLocator(ctx); repoLocator != nil && repoLocator.StatsdClient != nil {
        repoLocator.StatsdClient.Incr("api.users.get.cache_miss", nil, 1)
    }

    // データベースから取得
    user, err := uc.RUser.FindByID(ctx, id)
    return user, err
}
```

### 5. 処理時間の計測

```go
func (uc *UserUseCase) GetAllUsers(ctx context.Context) ([]*entities.User, error) {
    // 処理時間を計測
    start := time.Now()
    users, err := uc.RUser.FindAll(ctx)
    duration := time.Since(start)

    // Timingメトリクスを送信
    if repoLocator := appcontext.GetRepoLocator(ctx); repoLocator != nil && repoLocator.StatsdClient != nil {
        repoLocator.StatsdClient.Timing("api.users.list.duration", duration, nil, 1)
    }

    if err != nil {
        return nil, err
    }

    // ユーザー総数をGaugeで送信
    if repoLocator := appcontext.GetRepoLocator(ctx); repoLocator != nil && repoLocator.StatsdClient != nil {
        repoLocator.StatsdClient.Gauge("api.users.total", float64(len(users)), nil, 1)
    }

    return users, nil
}
```

---

## このプロジェクトで実装されているメトリクス

### 一覧表

| メトリクス名 | タイプ | 場所 | 説明 |
|------------|--------|------|------|
| `api.health.check` | Counter | HealthHandler | ヘルスチェックエンドポイントの呼び出し回数 |
| `api.users.create` | Counter | UserHandler | ユーザー作成試行回数（成功・失敗含む） |
| `api.users.create.success` | Counter | UserHandler | ユーザー作成成功回数 |
| `api.users.create.error` | Counter | UserHandler | ユーザー作成失敗回数 |
| `api.users.get.cache_hit` | Counter | UserUseCase | キャッシュヒット回数 |
| `api.users.get.cache_miss` | Counter | UserUseCase | キャッシュミス回数 |
| `api.users.list.duration` | Timing | UserUseCase | ユーザー一覧取得の処理時間 |
| `api.users.total` | Gauge | UserUseCase | 総ユーザー数 |

### メトリクス命名規則

このプロジェクトでは以下の命名規則を採用しています：

```
<namespace>.<resource>.<action>.<status>
```

**例**:
- `api.users.create` → API層、ユーザーリソース、作成アクション
- `api.users.create.success` → API層、ユーザーリソース、作成アクション、成功ステータス
- `api.users.get.cache_hit` → API層、ユーザーリソース、取得アクション、キャッシュヒット

---

## Datadogでの確認方法

### 1. Metrics Explorerでの確認

**手順**:
1. Datadog → **Metrics** → **Explorer**
2. 検索バーに `api.*` と入力
3. 実装したメトリクスが表示される

**表示されるメトリクス**:
- `api.health.check`
- `api.users.create`
- `api.users.create.success`
- `api.users.create.error`
- `api.users.get.cache_hit`
- `api.users.get.cache_miss`
- `api.users.list.duration`
- `api.users.total`

### 2. メトリクスの詳細確認

**Counterメトリクスの場合**:
```
api.users.create.count   → 累積値
api.users.create.rate    → 1秒あたりの発生回数（自動計算）
```

**Timingメトリクスの場合**:
```
api.users.list.duration.avg    → 平均処理時間
api.users.list.duration.p50    → 50パーセンタイル（中央値）
api.users.list.duration.p95    → 95パーセンタイル
api.users.list.duration.p99    → 99パーセンタイル
api.users.list.duration.max    → 最大処理時間
api.users.list.duration.min    → 最小処理時間
```

### 3. クエリ例

**ユーザー作成のエラー率**:
```
(sum:api.users.create.error.rate{*} / sum:api.users.create.rate{*}) * 100
```

**キャッシュヒット率**:
```
(sum:api.users.get.cache_hit.rate{*} /
 (sum:api.users.get.cache_hit.rate{*} + sum:api.users.get.cache_miss.rate{*})) * 100
```

**平均ユーザー作成成功率（過去1時間）**:
```
(sum:api.users.create.success.as_count{*}.rollup(sum, 3600) /
 sum:api.users.create.as_count{*}.rollup(sum, 3600)) * 100
```

---

## ダッシュボード作成

### ダッシュボードの設計

以下のウィジェットを含むダッシュボードを作成することを推奨します：

#### 1. リクエスト数

**ウィジェット**: Timeseries
**メトリクス**: `sum:api.users.create.rate{*}`
**説明**: 1秒あたりのユーザー作成リクエスト数

#### 2. エラー率

**ウィジェット**: Query Value
**クエリ**:
```
(sum:api.users.create.error.rate{*} / sum:api.users.create.rate{*}) * 100
```
**説明**: ユーザー作成のエラー率（%）

#### 3. キャッシュヒット率

**ウィジェット**: Query Value
**クエリ**:
```
(sum:api.users.get.cache_hit.rate{*} /
 (sum:api.users.get.cache_hit.rate{*} + sum:api.users.get.cache_miss.rate{*})) * 100
```
**説明**: キャッシュヒット率（%）

#### 4. ユーザー一覧取得の処理時間

**ウィジェット**: Timeseries
**メトリクス**:
- `avg:api.users.list.duration.avg{*}`
- `avg:api.users.list.duration.p95{*}`
- `avg:api.users.list.duration.p99{*}`

#### 5. 総ユーザー数

**ウィジェット**: Query Value
**メトリクス**: `avg:api.users.total{*}`

#### 6. 成功 vs エラー

**ウィジェット**: Stacked Area
**メトリクス**:
- `sum:api.users.create.success.rate{*}`
- `sum:api.users.create.error.rate{*}`

### ダッシュボードJSON例

```json
{
  "title": "User API Metrics",
  "widgets": [
    {
      "definition": {
        "type": "timeseries",
        "requests": [
          {
            "q": "sum:api.users.create.rate{*}",
            "display_type": "line"
          }
        ],
        "title": "User Creation Rate"
      }
    },
    {
      "definition": {
        "type": "query_value",
        "requests": [
          {
            "q": "(sum:api.users.create.error.rate{*} / sum:api.users.create.rate{*}) * 100",
            "aggregator": "avg"
          }
        ],
        "title": "Error Rate (%)"
      }
    }
  ]
}
```

---

## ベストプラクティス

### 1. メトリクス命名規則

**良い例**:
```
api.users.create
api.users.create.success
api.users.create.error
api.cache.hit
api.cache.miss
```

**悪い例**:
```
userCreate          // 曖昧
CreateUser          // 大文字を使わない
api_users_create    // アンダースコアではなくドットを使う
```

### 2. タグの活用

**基本的なタグ**:
```go
tags := []string{
    "endpoint:/api/users",
    "method:POST",
    "status:success",
}
statsdClient.Incr("api.requests", tags, 1)
```

**ビジネスタグ**:
```go
tags := []string{
    "user_type:premium",
    "region:us-east-1",
    "payment_method:credit_card",
}
statsdClient.Incr("api.orders.create", tags, 1)
```

### 3. メトリクスの粒度

**適切な粒度**:
- ユーザー作成成功・失敗を分ける
- エンドポイントごとにメトリクスを分ける

**粒度が細かすぎる例（避けるべき）**:
```go
// ユーザーIDごとにメトリクスを送信（カーディナリティが高すぎる）
statsdClient.Incr(fmt.Sprintf("api.users.%d.login", userID), nil, 1) // NG
```

### 4. サンプリングレート

高頻度のメトリクスにはサンプリングを使用：

```go
// 1%のリクエストのみメトリクスを送信
statsdClient.Incr("api.requests.high_frequency", nil, 0.01)
```

### 5. エラーハンドリング

```go
// nil チェックを必ず行う
if repoLocator != nil && repoLocator.StatsdClient != nil {
    if err := repoLocator.StatsdClient.Incr("api.users.create", nil, 1); err != nil {
        // エラーをログに記録するが、アプリケーションは継続
        logger.WithError(err).Warn("Failed to send metric")
    }
}
```

### 6. パフォーマンスへの影響

**UDP通信の特性**:
- 非ブロッキング（アプリケーションをブロックしない）
- メトリクス送信失敗してもアプリケーションに影響しない
- 軽量で高速

**注意点**:
- 大量のメトリクスを送信する場合はバッチ処理を検討
- DogStatsD Clientのバッファリング機能を活用

### 7. メトリクスのテスト

開発環境でメトリクスが正しく送信されているか確認：

```bash
# Datadog Agentのステータス確認
docker-compose exec datadog agent status

# DogStatsD統計確認
docker-compose exec datadog agent status | grep -A 10 "DogStatsD"
```

---

## トラブルシューティング

### 問題1: メトリクスがDatadogに表示されない

**原因と解決策**:

1. **Datadog Agentが起動していない**
   ```bash
   docker-compose ps datadog-agent
   docker-compose logs datadog-agent
   ```

2. **UDP通信がブロックされている**
   ```bash
   # Dockerネットワークの確認
   docker network inspect datadog-tour_default
   ```

3. **API Keyが設定されていない**
   ```bash
   # .envファイルを確認
   cat .env | grep DD_API_KEY
   ```

4. **メトリクス送信に時間がかかる**
   - メトリクスは数分後に表示される（リアルタイムではない）
   - 最低でも2-3分は待つ

### 問題2: メトリクスのカーディナリティが高すぎる

**原因**: タグの値が多すぎる（例: ユーザーID、IPアドレス）

**解決策**:
```go
// 悪い例
tags := []string{fmt.Sprintf("user_id:%d", userID)} // NG

// 良い例
tags := []string{"user_type:premium"} // OK
```

### 問題3: statsdClient.Incr()がエラーを返す

**原因**: StatsdClientがnilまたは初期化に失敗

**解決策**:
```go
// nil チェックを必ず行う
if repoLocator != nil && repoLocator.StatsdClient != nil {
    repoLocator.StatsdClient.Incr("api.users.create", nil, 1)
}
```

---

## まとめ

### Custom Metricsの重要ポイント

1. **DogStatsDを使用**: UDP通信で軽量・非同期
2. **適切なメトリクスタイプを選択**: Counter、Gauge、Timing等
3. **命名規則を統一**: `namespace.resource.action.status`
4. **タグを活用**: メトリクスをセグメント化
5. **nil チェック**: StatsdClientがnilでないことを確認
6. **エラーハンドリング**: メトリクス送信失敗でもアプリケーションは継続

### このプロジェクトで実装されているメトリクス

- `api.health.check`: ヘルスチェック回数
- `api.users.create`: ユーザー作成試行回数
- `api.users.create.success`: 成功回数
- `api.users.create.error`: 失敗回数
- `api.users.get.cache_hit`: キャッシュヒット回数
- `api.users.get.cache_miss`: キャッシュミス回数
- `api.users.list.duration`: 処理時間
- `api.users.total`: 総ユーザー数

### 次のステップ

1. Datadog Metrics Explorerでメトリクスを確認
2. ダッシュボードを作成してメトリクスを可視化
3. アラートを設定（エラー率、処理時間等）
4. 追加のビジネスメトリクスを実装

---

## 参考資料

- [Datadog DogStatsD Documentation](https://docs.datadoghq.com/developers/dogstatsd/)
- [datadog-go GitHub](https://github.com/DataDog/datadog-go)
- [Datadog Custom Metrics Guide](https://docs.datadoghq.com/metrics/custom_metrics/)
- [本プロジェクトの実装例](../../internal/)
