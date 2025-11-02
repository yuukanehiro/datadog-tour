# Datadog Tour - 実践的学習環境

Golang、MySQL、Redisを使用したDatadog APM学習用のサンプルアプリケーションです。

## 概要

このプロジェクトは、Datadogの主要機能を実践的に学ぶための環境を提供します：

- **APM (Application Performance Monitoring)**: 分散トレーシング
- **カスタムメトリクス**: DogStatsDを使用したビジネスメトリクス
- **ログ管理**: 構造化ログとトレースの相関
- **インフラストラクチャ監視**: MySQL、Redisのモニタリング
- **Continuous Profiler**: CPU/メモリプロファイリング

## アーキテクチャ

```
┌─────────────────┐
│  Datadog Agent  │
│   (APM/Logs)    │
└────────┬────────┘
         │
┌────────┴────────┐
│   Golang API    │
│    (Port 8080)  │
└────┬───────┬────┘
     │       │
     │       └──────────┐
┌────┴────┐      ┌─────┴─────┐
│  MySQL  │      │   Redis   │
│ (3306)  │      │   (6379)  │
└─────────┘      └───────────┘
```

## 前提条件

- Docker & Docker Compose
- Datadogアカウント (14日間無料トライアル)
- make コマンド (オプション)

## セットアップ手順

### 1. Datadog APIキーの取得

1. [Datadog](https://www.datadoghq.com/)にサインアップ
2. [API Keys ページ](https://app.datadoghq.com/organization-settings/api-keys)でAPIキーを取得

### 2. 環境変数の設定

```bash
# .env.exampleをコピー
cp .env.example .env

# .envファイルを編集してAPIキーを設定
# DD_API_KEY=your_actual_api_key_here
```

### 3. アプリケーションの起動

```bash
# すべてのサービスを起動
make up

# または
docker-compose up -d
```

### 4. 動作確認

```bash
# ヘルスチェック
curl http://localhost:8080/health

# すべてのAPIエンドポイントをテスト
make test-api
```

## API エンドポイント

### ヘルスチェック

```bash
GET /health
```

### ユーザー管理

```bash
# ユーザー作成
POST /api/users
Content-Type: application/json

{
  "name": "John Doe",
  "email": "john@example.com"
}

# ユーザー一覧取得
GET /api/users

# 特定のユーザー取得
GET /api/users/{id}
```

### キャッシュ操作

```bash
# キャッシュ設定
POST /api/cache/set
Content-Type: application/json

{
  "key": "my-key",
  "value": "my-value"
}

# キャッシュ取得
GET /api/cache/get/{key}
```

### テスト用エンドポイント

```bash
# 遅いエンドポイント (2秒待機)
GET /api/slow

# エラーエンドポイント
GET /api/error
```

## Datadog で確認できる内容

### 1. APM トレース

- サービスマップ: `api`, `mysql`, `redis` の依存関係
- トレース詳細: 各リクエストのレイテンシー
- エラー追跡: `/api/error` エンドポイントのエラー

**確認方法**: [APM > Services](https://app.datadoghq.com/apm/services)

### 2. カスタムメトリクス

- `api.health.check`: ヘルスチェックの呼び出し回数
- `api.users.create`: ユーザー作成回数
- `api.users.create.success`: 成功したユーザー作成
- `api.users.create.error`: 失敗したユーザー作成
- `api.users.total`: 総ユーザー数
- `api.users.list.duration`: ユーザー一覧取得の所要時間
- `api.users.get.cache_hit`: キャッシュヒット数
- `api.users.get.cache_miss`: キャッシュミス数

**確認方法**: [Metrics > Explorer](https://app.datadoghq.com/metric/explorer)

### 3. ログ

- 構造化されたJSONログ
- トレースIDとスパンIDの自動注入
- エラーログとトレースの相関

**確認方法**: [Logs > Explorer](https://app.datadoghq.com/logs)

### 4. インフラストラクチャ

- コンテナメトリクス
- MySQL パフォーマンスメトリクス
- Redis パフォーマンスメトリクス

**確認方法**: [Infrastructure > Containers](https://app.datadoghq.com/containers)

### 5. Continuous Profiler

- CPU使用率のプロファイル
- メモリアロケーションのプロファイル
- ホットスポットの特定

**確認方法**: [APM > Profiling](https://app.datadoghq.com/profiling)

## 学習演習

### Day 1-2: APM基礎

1. サービスマップを確認
2. `/api/users` にリクエストを送信してトレースを確認
3. `/api/slow` を呼び出して遅いリクエストを特定
4. `/api/error` を呼び出してエラートレースを確認

```bash
# 複数のリクエストを生成
for i in {1..10}; do
  curl -X POST http://localhost:8080/api/users \
    -H "Content-Type: application/json" \
    -d "{\"name\":\"User $i\",\"email\":\"user$i@example.com\"}"
done
```

### Day 3-4: ログ管理

1. Log Explorerでログを検索
2. トレースIDでログをフィルタリング
3. エラーログからトレースに移動
4. ログパターンの分析

### Day 5-6: カスタムメトリクス

1. Metrics Explorerで `api.*` メトリクスを検索
2. ダッシュボードを作成してメトリクスを可視化
3. キャッシュヒット率を計算: `cache_hit / (cache_hit + cache_miss)`

### Day 7-8: パフォーマンス最適化

1. Continuous Profilerでボトルネックを特定
2. `/api/users/{id}` のキャッシュ効果を確認
3. データベースクエリのN+1問題を探す

### Day 9-10: アラート設定

モニターを作成:

- エラー率が5%を超えた場合
- API応答時間が500msを超えた場合
- キャッシュミス率が80%を超えた場合

### Day 11-12: ダッシュボード作成

以下を含むダッシュボードを作成:

- リクエスト/秒
- 平均レスポンスタイム
- エラー率
- キャッシュヒット率
- アクティブユーザー数
- MySQL/Redisメトリクス

### Day 13-14: 本番運用シミュレーション

1. 負荷テストを実行
2. インシデント検出から解決までのフロー確認
3. SLOの設定 (例: 99.9% availability, p95 < 200ms)

## 便利なコマンド

```bash
# サービスの状態確認
make status

# ログ確認
make logs
make logs-api
make logs-datadog

# MySQL CLI
make mysql-cli

# Redis CLI
make redis-cli

# APIテスト
make test-api

# 再起動
make restart

# クリーンアップ
make clean
```

## 負荷テスト

Apache Bench (ab) を使用した負荷テスト:

```bash
# インストール (macOS)
brew install apache-bench

# 負荷テスト実行
ab -n 1000 -c 10 http://localhost:8080/api/users

# POST リクエストの負荷テスト
ab -n 100 -c 5 -p user.json -T application/json http://localhost:8080/api/users
```

user.json:
```json
{"name":"Load Test User","email":"loadtest@example.com"}
```

## トラブルシューティング

### Datadog Agentに接続できない

```bash
# Datadog Agentのログを確認
make logs-datadog

# APIキーが正しく設定されているか確認
docker-compose exec datadog agent status
```

### MySQLに接続できない

```bash
# MySQLのログを確認
docker-compose logs mysql

# MySQL CLIで接続テスト
make mysql-cli
```

### アプリケーションが起動しない

```bash
# アプリケーションログを確認
make logs-api

# コンテナを再ビルド
make build
```

## ディレクトリ構成

```
.
├── cmd/
│   └── api/
│       └── main.go         # アプリケーションのメインコード
├── docker/
│   ├── Dockerfile          # Golang アプリケーションのDockerfile
│   └── docker-compose.yml  # Docker Compose設定
├── docs/
│   └── datadog-14days-curriculum.md  # 学習カリキュラム
├── init/
│   └── init.sql            # MySQLの初期化スクリプト
├── go.mod                  # Go依存関係
├── go.sum                  # Go依存関係チェックサム
├── Makefile                # 便利なコマンド集
├── .env.example            # 環境変数のサンプル
├── .gitignore              # Git除外設定
└── README.md               # このファイル
```

## 実装されているDatadog機能

### APM Tracing

- `dd-trace-go`: Golangアプリケーションのトレーシング
- `gorilla/mux`: HTTPルーターの自動計装
- `database/sql`: MySQLクエリのトレーシング
- `go-redis`: Redisコマンドのトレーシング
- カスタムスパン: 特定の処理のトレーシング

### カスタムメトリクス

- `datadog-go/statsd`: DogStatsDクライアント
- Counter: イベント数のカウント
- Gauge: 現在の値の記録
- Timing: 処理時間の記録

### ログ管理

- `logrus`: 構造化ログライブラリ
- JSONフォーマット
- トレースID/スパンIDの自動注入
- ログレベルの設定

### Continuous Profiler

- CPUプロファイリング
- ヒーププロファイリング
- パフォーマンスボトルネックの特定

## 参考リソース

- [Datadog Documentation](https://docs.datadoghq.com/)
- [dd-trace-go GitHub](https://github.com/DataDog/dd-trace-go)
- [Datadog APM Guide](https://docs.datadoghq.com/tracing/)
- [カリキュラム](./docs/datadog-14days-curriculum.md)

## ライセンス

MIT License

## 貢献

プルリクエストを歓迎します！
