# Error Notification Setup Guide

このガイドでは、Datadogを使用した`error.notify`ベースのエラー通知システムのセットアップと運用方法について説明します。

## 目次

1. [概要](#概要)
2. [アーキテクチャ](#アーキテクチャ)
3. [error.notifyフィールド](#errornotifyフィールド)
4. [Terraformによるデプロイ](#terraformによるデプロイ)
5. [Slack通知設定](#slack通知設定)
6. [動作確認](#動作確認)
7. [トラブルシューティング](#トラブルシューティング)

---

## 概要

このシステムは、アプリケーションのエラーを`error.notify`フィールドで分類し、重要なシステムエラーのみSlackに通知します。

### 目的

- **アラート疲れの軽減**: 想定内エラー（バリデーションエラーなど）は通知しない
- **重要なエラーの即座検知**: システムエラーのみアラート対象
- **Infrastructure as Code**: Terraformで管理

### 構成要素

1. **アプリケーション側**: `error.notify`フィールドをログに追加
2. **Datadog Logs Pipeline**: エラーをカテゴリ分け
3. **Datadog Monitor**: `error.notify:true`のエラーを監視
4. **Slack Webhook**: アラート通知

---

## アーキテクチャ

```
Application (Go)
    ↓
    ├─ LogErrorWithTrace()          → error.notify: true  (アラート対象)
    └─ LogErrorWithTraceNotNotify() → error.notify: false (アラート対象外)
    ↓
Datadog Agent (ログ収集)
    ↓
Datadog Logs Pipeline
    ↓
    ├─ @error.notify:true  → alert.type: system_alert
    └─ @error.notify:false → alert.type: expected_error
    ↓
Datadog Monitor
    ↓
    Query: status:error @error.notify:true
    ↓
Slack Webhook Integration
    ↓
Slack Channel (#test)
```

---

## error.notifyフィールド

### 実装場所

`internal/common/logging/logger.go`

### 関数の使い分け

#### LogErrorWithTrace (error.notify: true)

**用途**: システムレベルのエラー

```go
// データベース接続エラー
logging.LogErrorWithTrace(ctx, logger, "usecase", "Database connection failed", err, logrus.Fields{
    "error.type": "system_error",
    "db.host":    "mysql.example.com",
})
```

**例**:
- データベース接続エラー
- 外部API呼び出し失敗
- メモリ不足
- ファイルIO失敗

#### LogErrorWithTraceNotNotify (error.notify: false)

**用途**: 想定内のエラー

```go
// バリデーションエラー
logging.LogErrorWithTraceNotNotify(ctx, logger, "usecase", "User already exists", err, logrus.Fields{
    "error.type": "validation_error",
    "user.email": "duplicate@example.com",
})
```

**例**:
- バリデーションエラー
- 重複エラー
- 権限エラー
- Not Foundエラー

### ログ出力形式

```json
{
  "message": "[handler] Unexpected error occurred",
  "level": "error",
  "error": {
    "notify": true,
    "type": "system_error"
  },
  "db": {
    "host": "mysql.example.com"
  },
  "dd.trace_id": "1234567890",
  "dd.span_id": "9876543210"
}
```

---

## Terraformによるデプロイ

### 前提条件

1. Datadog API Key
2. Datadog Application Key
3. Slack Webhook URL

### セットアップ

```bash
cd terraform

# 環境変数ファイルを作成
cp terraform.tfvars.example terraform.tfvars

# terraform.tfvarsを編集
# - datadog_api_key
# - datadog_app_key
# - slack_webhook_url
```

### デプロイ

```bash
# 初期化
terraform init

# 変更内容を確認
terraform plan

# デプロイ実行
terraform apply
```

### 作成されるリソース

#### 1. Logs Pipeline

**名前**: Error Notification Pipeline

**フィルタ**: `service:datadog-tour-api level:error`

**プロセッサ**:
- Category Processor: `error.notify`に基づいてカテゴリ分け
  - `@error.notify:true` → `alert.type: system_alert`
  - `@error.notify:false` → `alert.type: expected_error`

#### 2. System Error Alert Monitor

**クエリ**:
```
logs("service:datadog-tour-api status:error @error.notify:true")
  .index("*")
  .rollup("count")
  .last("5m") > 0
```

**閾値**: 5分間に1件以上のエラーでアラート

**通知先**: `@webhook-slack-test`

#### 3. High Error Rate Monitor

**クエリ**:
```
logs("service:datadog-tour-api status:error @error.notify:true")
  .index("*")
  .rollup("count")
  .last("15m") > 10
```

**閾値**:
- Warning: 15分間に5件以上
- Critical: 15分間に10件以上

**通知先**: `@webhook-slack-test`

#### 4. Webhook Integration

**名前**: slack-test

**URL**: Slack Webhook URL（terraform.tfvarsから取得）

**ペイロード**: Slack Attachment形式

---

## Slack通知設定

### Webhook URLの取得

1. Slack Workspaceにログイン
2. https://api.slack.com/apps を開く
3. アプリを作成または選択
4. **Incoming Webhooks** を有効化
5. チャンネルを選択してWebhook URLを取得

### terraform.tfvarsに設定

```hcl
slack_webhook_url = "https://hooks.slack.com/services/YOUR/WEBHOOK/URL"
slack_channel     = "@webhook-slack-test"
```

### 通知フォーマット

```
System Error Detected

Service: datadog-tour-api
Environment: development
Error Count: 3 errors in last 5 minutes

Details:
- Alert Type: System Alert (error.notify:true)
- Service: datadog-tour-api
- Error Type: system_error

Action Required:
Check Datadog Logs Explorer for details.
```

---

## 動作確認

### 1. システムエラーをテスト（アラート対象）

```bash
# アラートが発生する
curl http://localhost:8080/api/unexpected-error
```

**期待される動作**:
- ログに `error.notify: true` が記録される
- 約1-2分後にMonitorがAlert状態になる
- Slackに通知が届く

### 2. 想定内エラーをテスト（アラート対象外）

```bash
# アラートは発生しない
curl http://localhost:8080/api/expected-error
```

**期待される動作**:
- ログに `error.notify: false` が記録される
- Monitorはトリガーされない
- Slack通知は届かない

### 3. Datadog Logs Explorerで確認

**URL**: https://app.datadoghq.com/logs

**クエリ**:
```
service:datadog-tour-api status:error @error.notify:true
```

**確認項目**:
- `error.notify: true` が設定されているか
- `alert.type: system_alert` が追加されているか（Pipeline処理後）
- trace_id, span_id が記録されているか

### 4. Monitorの状態確認

**URL**: https://app.datadoghq.com/monitors/10387487

**確認項目**:
- Monitor状態（OK / Alert）
- 最後にトリガーされた時刻
- 評価されたログ数

---

## トラブルシューティング

### Slack通知が届かない

#### 1. Webhook URLの確認

```bash
# 手動でWebhook通知をテスト
curl -X POST YOUR_WEBHOOK_URL \
-H 'Content-Type: application/json' \
-d '{"text": "Test notification"}'
```

**期待される結果**: Slackに"Test notification"が表示される

#### 2. Monitor状態の確認

```bash
# MonitorがAlert状態か確認
# Datadog UI: https://app.datadoghq.com/monitors/10387487
```

#### 3. ログが記録されているか確認

```bash
# Logs Explorerでクエリ実行
# service:datadog-tour-api status:error @error.notify:true
```

### error.notifyフィールドが見つからない

#### 原因

アプリケーションコードが古い可能性があります。

#### 解決方法

```bash
# Dockerコンテナを再ビルド
cd /Users/kanehiroyuu/Documents/GitHub/datadog-tour
make rebuild-api

# または
docker-compose -f docker/docker-compose.yml --env-file .env up -d --build api
```

### MonitorがAlert状態にならない

#### 確認事項

1. **ログが届いているか**
   ```
   Query: service:datadog-tour-api status:error @error.notify:true
   ```

2. **Monitorのクエリが正しいか**
   - `@error.notify:true` (先頭に`@`が必要)
   - `status:error` (`level:error`ではない)

3. **評価期間が適切か**
   - System Error Alert: 5分間
   - エラー発生後、5分以内のデータが必要

#### Monitorを手動でテスト

Datadog UIで:
1. Monitor編集画面を開く
2. "..." メニュー → "Test Notifications"
3. テスト通知を送信

### Pipelineが動作しない

#### 確認事項

1. **Pipeline が有効か**
   - https://app.datadoghq.com/logs/pipelines
   - "Error Notification Pipeline"が有効になっているか

2. **フィルタが正しいか**
   - Filter: `service:datadog-tour-api level:error`

3. **新しいログでテスト**
   - Pipelineは新しいログにのみ適用される
   - 過去のログは再処理されない

---

## 参考リンク

### Datadog UI

- **Logs Explorer**: https://app.datadoghq.com/logs
- **Logs Pipelines**: https://app.datadoghq.com/logs/pipelines
- **Monitors**: https://app.datadoghq.com/monitors/manage
- **System Error Alert**: https://app.datadoghq.com/monitors/10387487
- **High Error Rate**: https://app.datadoghq.com/monitors/10387486
- **Webhook Integration**: https://app.datadoghq.com/integrations/webhooks

### Terraform

```bash
# 現在のリソースを確認
cd terraform
terraform show

# リソースを削除
terraform destroy

# 特定のリソースのみ削除
terraform destroy -target=datadog_monitor.system_error_alert
```

### ローカル環境

```bash
# エラーエンドポイント一覧
curl http://localhost:8080/api/slow              # 遅いリクエスト
curl http://localhost:8080/api/error             # エラー（旧）
curl http://localhost:8080/api/expected-error    # 想定内エラー (error.notify:false)
curl http://localhost:8080/api/unexpected-error  # システムエラー (error.notify:true)
curl http://localhost:8080/api/warn              # 警告ログ

# ログを確認
docker logs datadog-api -f

# コンテナ再起動
make restart
make rebuild-api
```

---

## まとめ

このシステムにより、以下を実現しています：

- **適切なアラート**: システムエラーのみ通知
- **Infrastructure as Code**: Terraformで管理
- **トレーサビリティ**: trace_id/span_idでログとトレースを紐付け
- **柔軟な拡張**: 新しいエラータイプや通知先を簡単に追加可能

---

最終更新日: 2025-11-05
