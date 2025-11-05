# Datadog Terraform Configuration

このディレクトリには、Datadog Logs Pipeline と Monitor を管理するためのTerraform設定が含まれています。

## 構成

### リソース

1. **Logs Pipeline** (`logs_pipeline.tf`)
   - エラーログを`error.notify`に基づいてカテゴリ分け
   - `alert.type:system_alert` - アラート対象
   - `alert.type:expected_error` - アラート対象外

2. **Monitors** (`monitor.tf`)
   - **System Error Alert**: 1件以上の`error.notify:true`エラーでSlack通知
   - **High Error Rate**: 15分間に10件以上のエラーでSlack通知

## セットアップ

### 1. Datadog APIキーの取得

1. [Datadog Organization Settings](https://app.datadoghq.com/organization-settings/api-keys) でAPI Keyを取得
2. [Datadog Application Keys](https://app.datadoghq.com/organization-settings/application-keys) でApplication Keyを作成

### 2. Slack連携の設定

1. Datadog UI → Integrations → Slack
2. Install Integration
3. 通知先のチャンネル名を確認（例: `@slack-alerts`）

### 3. 環境変数の設定

```bash
cp terraform.tfvars.example terraform.tfvars
```

`terraform.tfvars`を編集：

```hcl
datadog_api_key = "your-api-key-here"
datadog_app_key = "your-app-key-here"
datadog_api_url = "https://api.datadoghq.com"

service_name  = "datadog-tour-api"
environment   = "development"
slack_channel = "@slack-alerts"  # 自分のSlackチャンネル名
```

### 4. Terraformの初期化

```bash
cd terraform
terraform init
```

### 5. 変更内容の確認

```bash
terraform plan
```

### 6. デプロイ

```bash
terraform apply
```

確認のプロンプトで `yes` を入力。

## 動作確認

### エラーログをテスト

```bash
# アラート対象のエラー (error.notify: true)
curl http://localhost:8080/api/unexpected-error

# アラート対象外のエラー (error.notify: false)
curl http://localhost:8080/api/expected-error
```

### Datadogで確認

1. **Logs Explorer**: `service:datadog-tour-api level:error`
   - `alert.type:system_alert` 属性が追加されているか確認

2. **Monitors**: [Manage Monitors](https://app.datadoghq.com/monitors/manage)
   - 作成されたモニターが表示されているか確認

3. **Slack**: `/api/unexpected-error`を叩いた後、Slackに通知が届くか確認

## クリーンアップ

すべてのリソースを削除する場合：

```bash
terraform destroy
```

## ファイル構成

```
terraform/
├── provider.tf              # Datadog Provider設定
├── variables.tf             # 変数定義
├── logs_pipeline.tf         # Logs Pipeline設定
├── monitor.tf               # Monitor設定
├── terraform.tfvars.example # 環境変数のサンプル
├── terraform.tfvars         # 実際の環境変数（gitignore）
└── README.md                # このファイル
```

## 注意事項

- `terraform.tfvars`には機密情報が含まれるため、`.gitignore`に追加してください
- Datadog APIキーは安全に管理してください
- 本番環境では、環境変数やSecret Managerを使用することを推奨

## カスタマイズ

### モニターの閾値を変更

`monitor.tf`の`monitor_thresholds`を編集：

```hcl
monitor_thresholds {
  critical = 5   # エラー5件で通知
  warning  = 3   # エラー3件で警告
}
```

### 通知先を追加

`message`セクションに追加：

```hcl
message = <<-EOT
  ...
  ${var.slack_channel}
  @pagerduty-service
EOT
```

## トラブルシューティング

### API認証エラー

```
Error: error validating provider credentials
```

→ API KeyとApp Keyが正しいか確認

### Pipeline作成エラー

```
Error: error creating logs pipeline
```

→ Datadog UIで同名のPipelineが既に存在しないか確認
