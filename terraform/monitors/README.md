# Datadog Monitors - Terraform Configuration

このディレクトリには、Datadog Monitorsの設定をTerraformで管理するための設定ファイルが含まれています。

## ファイル構成

```
terraform/monitors/
├── main.tf              # Terraform & Provider設定
├── variables.tf         # 変数定義
├── error_rate.tf        # エラー率アラート
├── cache_hit_rate.tf    # キャッシュヒット率アラート
├── response_time.tf     # API応答時間アラート
├── service_down.tf      # サービスダウン検知
└── README.md            # このファイル
```

## 前提条件

### 1. Terraform のインストール

```bash
# macOS
brew install terraform

# 確認
terraform version
```

### 2. Datadog API Key と App Key の取得

1. [Datadog API Keys](https://app.datadoghq.com/organization-settings/api-keys) でAPI Keyを取得
2. [Datadog Application Keys](https://app.datadoghq.com/organization-settings/application-keys) でApp Keyを作成

**重要**: 使用しているDatadogリージョンを確認してください：
- URLが `app.datadoghq.com` → US1リージョン
- URLが `ap1.datadoghq.com` → AP1リージョン（日本）
- URLが `datadoghq.eu` → EU1リージョン

### 3. 環境変数の設定

```bash
# .env ファイルまたは ~/.zshrc に追加
export TF_VAR_datadog_api_key="your_api_key_here"
export TF_VAR_datadog_app_key="your_app_key_here"

# 反映
source ~/.zshrc
```

または、`terraform.tfvars` ファイルを作成：

```hcl
# terraform.tfvars
datadog_api_key = "your_api_key_here"
datadog_app_key = "your_app_key_here"
environment     = "production"
service_name    = "datadog-tour-api"
```

**重要**: `terraform.tfvars` をGitにコミットしないこと！

## 使い方

### 初期化

```bash
cd terraform/monitors
terraform init
```

### 変更内容の確認

```bash
terraform plan
```

**出力例**:
```
Terraform will perform the following actions:

  # datadog_monitor.user_creation_error_rate will be created
  + resource "datadog_monitor" "user_creation_error_rate" {
      + name    = "[datadog-tour-api] High User Creation Error Rate"
      + type    = "metric alert"
      ...
    }
```

### 適用

```bash
terraform apply
```

確認メッセージが表示されるので `yes` を入力。

### 削除

```bash
# 特定のMonitorを削除
terraform destroy -target=datadog_monitor.user_creation_error_rate

# すべてのMonitorを削除
terraform destroy
```

## カスタマイズ

### 環境変数でカスタマイズ

```bash
# Staging環境用に適用
terraform apply -var="environment=staging" -var="service_name=datadog-tour-api-staging"
```

### variables.tf の編集

デフォルト値を変更する場合は `variables.tf` を編集：

```hcl
variable "slack_channel" {
  description = "Slack channel for alerts"
  type        = string
  default     = "@your-slack-channel"  # ここを変更
}
```

### 閾値の変更

各Monitorファイル（`error_rate.tf` 等）の `monitor_thresholds` を編集：

```hcl
monitor_thresholds {
  critical = 15  # 10 → 15 に変更
  warning  = 8   # 5 → 8 に変更
}
```

## 設定されるMonitor一覧

| Monitor名 | ファイル | 説明 | 閾値 | Monitor ID |
|----------|---------|------|------|-----------|
| High User Creation Error Rate | error_rate.tf | ユーザー作成エラー率 | Warning: 5%, Critical: 10% | 10529855 |
| High Cache Miss Count | cache_hit_rate.tf | キャッシュミス数 | Warning: 50, Critical: 100 | 10529857 |
| Slow API Response Time | response_time.tf | API応答時間（p95） | Warning: 500ms, Critical: 1000ms | 10529858 |
| Service Down | service_down.tf | ヘルスチェック失敗 | Critical: 1分間に0回以下 | 10529856 |

**注**: Monitor IDは2025-11-17に作成された本番環境のものです。

## Slack通知の設定

### 1. Datadog Slack Integration の設定

1. Datadog → **Integrations** → **Slack**
2. **Add Slack Account**
3. チャネルを選択してインストール

### 2. Slack Channel名の確認

インストールしたチャネル名を確認（例: `#datadog-alerts`）

### 3. variables.tf で設定

```hcl
variable "slack_channel" {
  default = "@slack-datadog-alerts"  # #datadog-alerts → @slack-datadog-alerts
}

variable "critical_slack_channel" {
  default = "@slack-datadog-critical"
}
```

## PagerDuty通知の設定

### 1. Datadog PagerDuty Integration の設定

1. Datadog → **Integrations** → **PagerDuty**
2. PagerDuty Integration Key を入力

### 2. Service名の確認

PagerDutyで作成したService名を確認

### 3. variables.tf で設定

```hcl
variable "pagerduty_service" {
  default = "@pagerduty-oncall"
}
```

## トラブルシューティング

### 問題1: Monitor Status が "NO DATA"

**症状**:
- モニターは正常に作成されている
- Datadog UIでモニターのステータスが「NO DATA」と表示される
- メトリクス自体はDatadogに送信されている

**原因**:
1. **`require_full_window = true`（デフォルト値）**: 評価期間の完全なデータウィンドウが必要
   - 例: `last_5m`の場合、連続した5分間のデータが必要
   - 新規作成したモニターや不定期なメトリクスではデータが揃わない
2. **メトリクスの送信頻度が低い**: データポイントの間隔が空いている
3. **評価タイミング**: モニターが評価される前にデータが届いていない

**解決方法**:

**1. `require_full_window = false`に設定（推奨）**:

全てのモニターファイルに以下を追加：

```hcl
resource "datadog_monitor" "example" {
  # ... 他の設定 ...

  notify_no_data      = false
  renotify_interval   = 0
  notify_audit        = false
  timeout_h           = 0
  include_tags        = true
  require_full_window = false  # ← この行を追加

  tags = [...]
}
```

適用：
```bash
cd terraform/monitors
terraform apply -auto-approve
```

**2. メトリクスを継続的に生成**:

```bash
# APIを定期的に叩く（1分間）
for i in {1..10}; do
  curl -s http://localhost:8080/health > /dev/null
  curl -s http://localhost:8080/api/users > /dev/null
  sleep 6
done
```

**3. 確認手順**:

```bash
# 1. Metrics Explorerでメトリクスが存在するか確認
open "https://ap1.datadoghq.com/metric/explorer"
# メトリクス名で検索: api.health.check, api.users.create など

# 2. 数分待つ（モニターは1分ごとに評価される）

# 3. Monitors画面でステータスを確認
open "https://ap1.datadoghq.com/monitors/manage"
```

**デバッグ方法**:

```bash
# メトリクスが送信されているか確認
curl -X GET "https://api.ap1.datadoghq.com/api/v1/metrics?from=1763309000" \
  -H "DD-API-KEY: your_api_key" \
  -H "DD-APPLICATION-KEY: your_app_key" | \
  jq -r '.metrics[]' | grep api

# 期待される出力:
# api.health.check
# api.users.create
# api.users.create.error
# ...
```

### 問題2: 403 Forbidden エラー

**エラー**:
```
Error: 403 Forbidden
error validating monitor from /api/v1/monitor/validate: 403 Forbidden
```

**原因**: APIエンドポイントがDatadogリージョンと一致していない

**解決策**:

`main.tf`でリージョンに応じたAPIエンドポイントを設定：

```hcl
provider "datadog" {
  api_key = var.datadog_api_key
  app_key = var.datadog_app_key
  api_url = "https://api.ap1.datadoghq.com"  # AP1リージョン（日本）
}
```

主要なリージョンエンドポイント：
- US1（デフォルト）: `https://api.datadoghq.com`
- AP1（日本）: `https://api.ap1.datadoghq.com`
- EU1: `https://api.datadoghq.eu`

API keyが正しいか確認：
```bash
# AP1リージョンの場合
curl -X GET "https://api.ap1.datadoghq.com/api/v1/validate" \
  -H "DD-API-KEY: your_api_key"
```

### 問題3: Invalid query エラー

**エラー**:
```
Error: error validating monitor from /api/v1/monitor/validate: 400 Bad Request:
{"errors":["The value provided for parameter 'query' is invalid"]}
```

**原因**: クエリ文字列に比較演算子と閾値が含まれていない

**解決策**:

クエリに`> 10`や`<= 0`などの比較演算子を**必ず含める**：

```hcl
# ❌ 間違い
query = "avg(last_5m):avg:api.users.create{*}"

# ✅ 正しい
query = "avg(last_5m):avg:api.users.create{*} > 10"
```

その他のクエリの注意点：
- `.as_count`や`.rollup()`は使用できない（シンプルな形式を使用）
- パーセンタイル指定は`.p95`ではなく`.95percentile`
- 環境フィルタリングは慎重に（メトリクスが存在しない場合エラー）

### 問題4: Provider authentication failed

**エラー**:
```
Error: error validating provider credentials: invalid API/APP key combination
```

**解決策**:
```bash
# API KeyとApp Keyを確認
echo $TF_VAR_datadog_api_key
echo $TF_VAR_datadog_app_key

# 正しく設定されていない場合は再設定
export TF_VAR_datadog_api_key="your_correct_api_key"
export TF_VAR_datadog_app_key="your_correct_app_key"
```

### 問題5: Monitor already exists

**エラー**:
```
Error: A monitor with the name already exists
```

**解決策**:

既存のMonitorをインポート：
```bash
# Monitor IDを確認（Datadog UI で確認）
terraform import datadog_monitor.user_creation_error_rate 12345678
```

または、既存のMonitorを削除してから適用。

### 問題6: Metrics not found

**エラー**:
```
Error: Metric 'api.users.create' not found
```

**解決策**:

1. アプリケーションが実際に稼働していて、メトリクスを送信しているか確認
2. Datadog Metrics Explorerでメトリクス名を確認
3. メトリクスが生成されてから反映まで数分かかる場合がある

```bash
# APIを叩いてメトリクスを生成
curl http://localhost:8080/api/users
curl -X POST http://localhost:8080/api/users \
  -H "Content-Type: application/json" \
  -d '{"name":"Test","email":"test@example.com"}'
```

### デバッグ方法

```bash
# Terraformのデバッグモードを有効化
export TF_LOG=DEBUG
terraform plan

# 特定のリソースのみ適用
terraform apply -target=datadog_monitor.service_down

# Terraform stateの確認
terraform state list
terraform state show datadog_monitor.user_creation_error_rate
```

## ベストプラクティス

### 1. 環境ごとに分ける

```bash
# ディレクトリ構成
terraform/monitors/
├── production/
│   ├── main.tf
│   └── ...
└── staging/
    ├── main.tf
    └── ...
```

### 2. State管理

本番環境ではリモートStateを使用：

```hcl
# backend.tf
terraform {
  backend "s3" {
    bucket = "my-terraform-state"
    key    = "datadog/monitors/terraform.tfstate"
    region = "us-east-1"
  }
}
```

### 3. Moduleの活用

共通設定をModuleとして切り出す：

```hcl
# modules/monitor/main.tf
resource "datadog_monitor" "this" {
  name    = var.name
  type    = var.type
  message = var.message
  query   = var.query

  monitor_thresholds {
    critical = var.critical_threshold
    warning  = var.warning_threshold
  }

  tags = var.tags
}
```

### 4. タグの統一

すべてのMonitorに共通のタグを設定：

```hcl
locals {
  common_tags = [
    "service:${var.service_name}",
    "env:${var.environment}",
    "managed_by:terraform"
  ]
}

resource "datadog_monitor" "example" {
  tags = concat(local.common_tags, ["severity:high"])
}
```

## 重要なポイントまとめ

### 1. require_full_window = false を設定（最重要）

**必ず全てのモニターに設定してください**:

```hcl
resource "datadog_monitor" "example" {
  # ...
  require_full_window = false  # これがないとNO DATAになる
}
```

**理由**:
- デフォルトは`true`で、評価期間の完全なデータウィンドウが必要
- `last_5m`の場合、連続した5分間のデータが揃うまで「NO DATA」
- `false`に設定すると部分的なデータでも評価可能

### 2. リージョン設定は必須

`main.tf`でDatadogリージョンに応じたAPIエンドポイントを明示的に設定すること。

### 3. クエリには比較演算子が必須

```hcl
# ❌ これではエラー
query = "avg(last_5m):avg:api.users.create{*}"

# ✅ 必ず比較演算子を含める
query = "avg(last_5m):avg:api.users.create{*} > 10"
```

### 4. メトリクス名は正確に

- Datadog Metrics Explorerで実際のメトリクス名を確認
- `.as_count`や`.rollup()`は使えない
- パーセンタイルは`.95percentile`形式

### 5. まずメトリクスを生成

Monitorを作成する前に、アプリケーションを実行してメトリクスが実際に送信されているか確認。

## 参考資料

- [Terraform Datadog Provider Documentation](https://registry.terraform.io/providers/DataDog/datadog/latest/docs)
- [Datadog Monitor API](https://docs.datadoghq.com/api/latest/monitors/)
- [Datadog Regions](https://docs.datadoghq.com/getting_started/site/)
- [本プロジェクトのアラート設定ガイド](../../docs/custom-metrics/alerts-guide.md)
- [本プロジェクトのCustom Metricsガイド](../../docs/custom-metrics/custom-metrics-guide.md)
