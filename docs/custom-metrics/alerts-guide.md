# Custom Metricsを利用したアラート設定ガイド

## 目次
1. [アラートの基本概念](#アラートの基本概念)
2. [Monitor（モニター）の種類](#monitorモニターの種類)
3. [実装されているメトリクスに対する推奨アラート](#実装されているメトリクスに対する推奨アラート)
4. [アラート設定手順](#アラート設定手順)
5. [アラート設定例](#アラート設定例)
6. [通知チャネルの設定](#通知チャネルの設定)
7. [ベストプラクティス](#ベストプラクティス)
8. [Terraform による自動化](#terraformによる自動化)

---

## アラートの基本概念

### Monitorとは

Datadogの**Monitor（モニター）**は、メトリクスやログを監視し、閾値を超えた場合に通知を送る仕組みです。

### アラートのライフサイクル

```
Normal (正常)
    ↓
    条件を満たす
    ↓
Alert (警告) → 通知送信
    ↓
    条件が解消
    ↓
Recovered (回復) → 回復通知送信
    ↓
Normal (正常)
```

### アラートの重要性

1. **問題の早期発見**: エラー率が上昇したらすぐに気づける
2. **SLO達成**: サービスレベル目標の維持
3. **ビジネスインパクトの最小化**: 問題が拡大する前に対処
4. **自動化**: 人間が常時監視する必要がない

---

## Monitor（モニター）の種類

### 1. Metric Monitor

**用途**: メトリクスが閾値を超えた場合にアラート

**例**:
- エラー率が5%を超えた
- API応答時間が500msを超えた
- キャッシュヒット率が80%を下回った

### 2. Anomaly Monitor

**用途**: 過去のパターンと比較して異常を検知

**例**:
- リクエスト数が通常の2倍になった（トラフィックスパイク）
- 深夜に通常発生しないアクティビティが検出された

### 3. Outlier Monitor

**用途**: 複数のホストやサービス間で異常な挙動を検知

**例**:
- 1台のサーバーだけエラー率が高い
- 特定のリージョンでレイテンシが高い

### 4. Forecast Monitor

**用途**: 現在のトレンドから将来の値を予測してアラート

**例**:
- このままだとディスク容量が3日後に満杯になる
- ユーザー数の増加により来週中にデータベース容量が不足する

### 5. Composite Monitor

**用途**: 複数のMonitorを組み合わせた条件でアラート

**例**:
- エラー率が高い AND レスポンス時間も遅い
- キャッシュミス率が高い OR データベース接続エラーが発生

---

## 実装されているメトリクスに対する推奨アラート

このプロジェクトで実装されているメトリクスに対する推奨アラート設定です。

### 1. ユーザー作成エラー率が高い

**メトリクス**: `api.users.create.error` / `api.users.create`

**閾値**:
- Warning: エラー率 > 5%
- Critical: エラー率 > 10%

**理由**: ユーザー登録ができないと新規ユーザー獲得に影響

**設定**:
```
Alert when: (sum:api.users.create.error.as_count{*}.rollup(sum, 300) /
             sum:api.users.create.as_count{*}.rollup(sum, 300)) * 100 > 5
```

### 2. API応答時間が遅い

**メトリクス**: `api.users.list.duration.p95`

**閾値**:
- Warning: p95 > 500ms
- Critical: p95 > 1000ms

**理由**: ユーザー体験に直結、SEOにも影響

**設定**:
```
Alert when: avg:api.users.list.duration.p95{*} > 0.5 (seconds)
```

### 3. キャッシュヒット率が低い

**メトリクス**: `api.users.get.cache_hit` / (`api.users.get.cache_hit` + `api.users.get.cache_miss`)

**閾値**:
- Warning: ヒット率 < 80%
- Critical: ヒット率 < 60%

**理由**: キャッシュヒット率が低いとDB負荷が増加

**設定**:
```
Alert when: (sum:api.users.get.cache_hit.as_count{*}.rollup(sum, 300) /
            (sum:api.users.get.cache_hit.as_count{*}.rollup(sum, 300) +
             sum:api.users.get.cache_miss.as_count{*}.rollup(sum, 300))) * 100 < 80
```

### 4. ユーザー作成数が異常に多い

**メトリクス**: `api.users.create`

**タイプ**: Anomaly Monitor

**理由**: ボット攻撃やスパム登録の可能性

**設定**:
```
Alert when: api.users.create の増加率が通常の3倍を超える
```

### 5. ヘルスチェックが失敗している

**メトリクス**: `api.health.check`

**閾値**:
- Critical: 過去1分間にヘルスチェックが0回

**理由**: サービスがダウンしている可能性

**設定**:
```
Alert when: sum:api.health.check.as_count{*}.rollup(sum, 60) == 0
```

### 6. ユーザー総数の急激な減少

**メトリクス**: `api.users.total`

**タイプ**: Change Alert

**閾値**:
- Critical: 過去1時間で10%以上減少

**理由**: データ削除バグやデータベース障害の可能性

**設定**:
```
Alert when: change(avg:api.users.total{*}, 3600) < -10%
```

---

## アラート設定手順

### 方法1: Datadog UIから設定

#### Step 1: Monitorの作成

1. Datadog → **Monitors** → **New Monitor**
2. Monitor typeを選択（Metric、Anomaly等）

#### Step 2: メトリクスの選択

**エラー率のアラート例**:

1. **Define the metric**:
   ```
   (sum:api.users.create.error.as_count{*}.rollup(sum, 300) /
    sum:api.users.create.as_count{*}.rollup(sum, 300)) * 100
   ```

2. **Set alert conditions**:
   - Alert threshold: `> 5`
   - Warning threshold: `> 3`
   - Evaluation window: `last 5 minutes`

#### Step 3: 通知の設定

**件名**:
```
[{{#is_alert}}Alert{{/is_alert}}{{#is_warning}}Warning{{/is_warning}}] User Creation Error Rate is {{value}}%
```

**メッセージ**:
```
## User Creation Error Rate Alert

**Current Error Rate**: {{value}}%
**Threshold**: {{threshold}}%

### Details
- **Service**: datadog-tour-api
- **Environment**: {{env.name}}
- **Time**: {{last_triggered_at}}

### Metrics
- Total user creation attempts: {{api.users.create.as_count}}
- Failed attempts: {{api.users.create.error.as_count}}

### Actions
1. Check application logs for error details
2. Verify database connectivity
3. Check for recent deployments

### Links
- [View Logs](https://app.datadoghq.com/logs?query=service:datadog-tour-api%20error)
- [View APM Traces](https://app.datadoghq.com/apm/traces?query=service:datadog-tour-api%20error:true)
- [View Dashboard](https://app.datadoghq.com/dashboard/your-dashboard-id)

{{#is_alert}}
@slack-alerts @pagerduty
{{/is_alert}}

{{#is_recovery}}
The error rate has returned to normal levels.
{{/is_recovery}}
```

#### Step 4: 通知先の設定

通知先を追加:
- `@slack-alerts` - Slackチャネル
- `@pagerduty` - PagerDuty
- `@email-ops@example.com` - メールアドレス

#### Step 5: 保存

**Monitor name**: `[User API] High Error Rate`
**Tags**: `service:datadog-tour-api`, `team:backend`, `severity:high`

---

## アラート設定例

### 例1: エラー率アラート（詳細版）

**Monitor Type**: Metric Monitor

**Query**:
```
(sum:api.users.create.error.as_count{env:production}.rollup(sum, 300) /
 sum:api.users.create.as_count{env:production}.rollup(sum, 300)) * 100
```

**Conditions**:
- Alert threshold: `> 10`
- Warning threshold: `> 5`
- No data: Notify after 10 minutes
- Auto-resolve: After 15 minutes

**Notification**:
```
## High User Creation Error Rate

**Error Rate**: {{value}}%
**Environment**: {{env.name}}

### Impact
- Users cannot register
- Business impact: High
- Expected resolution time: 30 minutes

### Triage Steps
1. Check [Recent Deployments](https://app.datadoghq.com/apm/services/datadog-tour-api)
2. Review [Error Logs](https://app.datadoghq.com/logs?query=service:datadog-tour-api%20status:error)
3. Verify database connection: `docker-compose exec api curl http://localhost:8080/health`
4. Check MySQL status: `docker-compose logs mysql`

### Runbook
See: https://wiki.example.com/runbooks/user-creation-errors

{{#is_alert}}
@slack-critical @pagerduty-oncall
{{/is_alert}}
```

### 例2: キャッシュヒット率アラート

**Monitor Type**: Metric Monitor

**Query**:
```
(sum:api.users.get.cache_hit.as_count{*}.rollup(sum, 600) /
(sum:api.users.get.cache_hit.as_count{*}.rollup(sum, 600) +
 sum:api.users.get.cache_miss.as_count{*}.rollup(sum, 600))) * 100
```

**Conditions**:
- Warning threshold: `< 80`
- Critical threshold: `< 60`
- Evaluation window: `last 10 minutes`

**Notification**:
```
## Low Cache Hit Rate

**Cache Hit Rate**: {{value}}%
**Expected**: > 80%

### Impact
- Increased database load
- Slower API responses
- Potential database connection exhaustion

### Investigation
1. Check Redis status: `docker-compose logs redis`
2. Verify cache TTL settings
3. Check for cache key pattern changes
4. Monitor database query performance

### Metrics
- Cache hits: {{api.users.get.cache_hit.as_count}}
- Cache misses: {{api.users.get.cache_miss.as_count}}

{{#is_warning}}
@slack-alerts
{{/is_warning}}

{{#is_alert}}
@slack-critical @pagerduty-oncall
{{/is_alert}}
```

### 例3: API応答時間アラート

**Monitor Type**: Metric Monitor

**Query**:
```
avg:api.users.list.duration.p95{*}
```

**Conditions**:
- Warning threshold: `> 0.5` (500ms)
- Critical threshold: `> 1.0` (1000ms)
- Evaluation window: `last 5 minutes`

**Notification**:
```
## Slow API Response Time

**p95 Response Time**: {{value}} seconds
**Threshold**: {{threshold}} seconds

### User Impact
- Degraded user experience
- Potential timeout errors
- SEO impact

### Check
1. Database query performance
2. Cache hit rate: [View Dashboard](https://app.datadoghq.com/dashboard/cache-metrics)
3. CPU/Memory usage
4. Active database connections

### APM Traces
[View Slow Traces](https://app.datadoghq.com/apm/traces?query=service:datadog-tour-api%20resource_name:"GET%20/api/users"%20@duration:>500ms)

{{#is_alert}}
@slack-performance
{{/is_alert}}
```

### 例4: 異常なユーザー作成数（Anomaly Detection）

**Monitor Type**: Anomaly Monitor

**Query**:
```
avg:api.users.create.as_rate{*}
```

**Conditions**:
- Alert when: Deviates 3x from normal
- Algorithm: Agile
- Seasonality: None

**Notification**:
```
## Abnormal User Creation Activity

**Current Rate**: {{value}} users/sec
**Expected Range**: {{baseline_min}} - {{baseline_max}}

### Possible Causes
- Bot attack
- Marketing campaign
- Spam registration
- Application bug

### Immediate Actions
1. Review recent user registrations
2. Check for suspicious patterns (same email domain, IP)
3. Enable rate limiting if needed
4. Monitor for fraud indicators

### Metrics
[View User Creation Dashboard](https://app.datadoghq.com/dashboard/user-metrics)

{{#is_alert}}
@slack-security @slack-ops
{{/is_alert}}
```

### 例5: サービスダウン検知

**Monitor Type**: Metric Monitor

**Query**:
```
sum:api.health.check.as_count{*}.rollup(sum, 60)
```

**Conditions**:
- Critical threshold: `== 0`
- No data: Notify after 2 minutes

**Notification**:
```
## Service Down - No Health Check Response

**Status**: Service appears to be down
**Last Check**: {{last_triggered_at}}

### Critical Actions Required
1. Check application status: `docker-compose ps api`
2. Check application logs: `docker-compose logs api --tail 50`
3. Restart if needed: `make restart-api`
4. Verify all dependencies (MySQL, Redis, Datadog Agent)

### Escalation
If not resolved in 5 minutes, escalate to:
- Engineering Manager: @manager
- On-call Engineer: @pagerduty-oncall

{{#is_alert}}
@pagerduty-critical @slack-critical
{{/is_alert}}
```

---

## 通知チャネルの設定

### Slack統合

1. Datadog → **Integrations** → **Slack**
2. **Add Slack Account**
3. チャネルを選択してインストール
4. Monitorで `@slack-alerts` として使用

### PagerDuty統合

1. Datadog → **Integrations** → **PagerDuty**
2. PagerDuty Integration Keyを入力
3. Monitorで `@pagerduty` として使用

### Email通知

Monitorメッセージに直接メールアドレスを記載：
```
@email-ops@example.com
@email-team@example.com
```

### Webhook

カスタムWebhookを設定してSlack、Teams、独自システムに通知：

1. Datadog → **Integrations** → **Webhooks**
2. Webhook URLを設定
3. Monitorで `@webhook-custom` として使用

---

## ベストプラクティス

### 1. アラート疲れを避ける

**問題**: アラートが多すぎると無視されるようになる

**解決策**:
- 重要度に応じて通知チャネルを分ける
- 閾値を適切に設定（過去データを参考に）
- 自動回復するアラートはWarningレベルに

### 2. アラートの優先度

**Critical（重大）**:
- サービスダウン
- エラー率 > 10%
- データ損失の可能性

**Warning（警告）**:
- パフォーマンス低下
- エラー率 > 5%
- キャッシュヒット率低下

**Info（情報）**:
- 異常なトラフィックパターン
- リソース使用率の上昇

### 3. アラートメッセージの構成

必ず含めるべき情報：
1. **何が起きているか**: 具体的な問題
2. **影響範囲**: ユーザーへの影響
3. **対処手順**: 明確なアクションアイテム
4. **関連リンク**: ログ、トレース、ダッシュボード
5. **エスカレーション**: 誰に連絡すべきか

### 4. Runbookの準備

各アラートに対応するRunbookを用意：

```markdown
# Runbook: High User Creation Error Rate

## Symptoms
- Error rate > 5%
- Users cannot register

## Investigation Steps
1. Check application logs for specific error messages
2. Verify database connectivity
3. Check for recent code deployments
4. Review database query performance

## Common Causes
- Database connection pool exhausted
- Validation errors (duplicate email)
- Third-party API timeout
- Database schema changes

## Resolution Steps
1. If database issue: Restart MySQL container
2. If connection pool: Increase max connections
3. If validation: Check recent user data
4. If deployment: Rollback to previous version

## Prevention
- Add database connection pool monitoring
- Implement exponential backoff for retries
- Add input validation tests
```

### 5. アラートのテスト

本番環境でアラートが正しく動作するか定期的にテスト：

```bash
# エラーを意図的に発生させてアラートをテスト
curl -X POST http://localhost:8080/api/users \
  -H "Content-Type: application/json" \
  -d '{"name":"Test","email":"existing@example.com"}'  # 重複メール
```

### 6. アラートの定期レビュー

月次でアラート設定をレビュー：
- 不要なアラートを削除
- 閾値を調整
- 新しいメトリクスに対するアラート追加

---

## Terraformによる自動化

アラート設定をコードで管理（Infrastructure as Code）：

### ディレクトリ構成

```
terraform/monitors/
├── main.tf              # Provider設定
├── variables.tf         # 変数定義
├── terraform.tfvars     # 変数の値（APIキー等）
├── error_rate.tf        # エラー率モニター
├── cache_hit_rate.tf    # キャッシュミスモニター
├── response_time.tf     # レスポンスタイムモニター
├── service_down.tf      # サービスダウンモニター
└── README.md
```

### 重要な設定ポイント

#### 1. AP1リージョンの設定

日本リージョン（AP1）を使用する場合、`main.tf`でAPIエンドポイントを明示的に設定：

```hcl
provider "datadog" {
  api_key = var.datadog_api_key
  app_key = var.datadog_app_key
  api_url = "https://api.ap1.datadoghq.com"  # AP1リージョン
}
```

#### 2. クエリ文字列の書き方

**重要**: Datadog Terraformプロバイダーでは、クエリ文字列に**比較演算子と閾値を含める**必要があります。

**正しい例**:
```hcl
query = "avg(last_5m):avg:api.users.list.duration.95percentile{*} > 1.0"
```

**間違った例**:
```hcl
query = "avg(last_5m):avg:api.users.list.duration.95percentile{*}"
```

#### 3. メトリクス名の注意点

- `.as_count`や`.rollup()`などの関数は使用できない
- パーセンタイル指定は`.p95`ではなく`.95percentile`

#### 4. require_full_window の設定

**重要**: `require_full_window = false`を設定することを推奨します。

```hcl
resource "datadog_monitor" "example" {
  # ...
  require_full_window = false  # これを追加
}
```

**理由**:
- `require_full_window = true`（デフォルト）の場合、評価期間（`last_5m`など）の**完全なデータウィンドウ**が必要
- 新しく作成したモニターや、メトリクスの送信が不定期な場合、データが揃わず「NO DATA」になる
- `false`に設定すると、部分的なデータでも評価が可能になり、よりリアルタイムなアラートが実現できる

**例**:
```hcl
# 評価期間が last_5m の場合
require_full_window = true   # 5分間の完全なデータが必要（300秒分）
require_full_window = false  # 部分的なデータでも評価可能（例: 1分分のデータでも評価）
```

### 実装済みモニター

以下の4つのモニターが実装されています（2025-11-17作成）：

#### 1. error_rate.tf - ユーザー作成エラー率

```hcl
resource "datadog_monitor" "user_creation_error_rate" {
  name    = "[${var.service_name}] High User Creation Error Rate"
  type    = "metric alert"
  message = <<-EOT
    ## High User Creation Error Rate

    **Error Rate**: {{value}}%
    **Environment**: ${var.environment}
    **Service**: ${var.service_name}

    ### Impact
    - Users cannot register
    - Business impact: High
    - Expected resolution time: 30 minutes

    ### Triage Steps
    1. Check [Recent Deployments](https://app.datadoghq.com/apm/services/${var.service_name})
    2. Review [Error Logs](https://app.datadoghq.com/logs?query=service:${var.service_name}%20status:error)
    3. Verify database connection
    4. Check MySQL status

    {{#is_warning}}
    ${var.slack_channel}
    {{/is_warning}}

    {{#is_alert}}
    ${var.critical_slack_channel} ${var.pagerduty_service}
    {{/is_alert}}
  EOT

  query = "avg(last_5m):(sum:api.users.create.error{*} / sum:api.users.create{*}) * 100 > 10"

  monitor_thresholds {
    critical = 10
    warning  = 5
  }

  notify_no_data      = false
  renotify_interval   = 60
  notify_audit        = false
  timeout_h           = 0
  include_tags        = true
  require_full_window = false  # 部分的なデータでも評価

  tags = [
    "service:${var.service_name}",
    "env:${var.environment}",
    "team:backend",
    "severity:high",
    "type:error_rate"
  ]
}
```

**モニターID**: 10529855

#### 2. cache_hit_rate.tf - キャッシュミス数

```hcl
resource "datadog_monitor" "cache_miss_rate" {
  name    = "[${var.service_name}] High Cache Miss Count"
  type    = "metric alert"
  message = <<-EOT
    ## High Cache Miss Count

    **Cache Misses**: {{value}} in last 10 minutes
    **Environment**: ${var.environment}

    ### Impact
    - Increased database load
    - Slower API responses
    - Potential database connection exhaustion

    ### Investigation
    1. Check Redis status
    2. Verify cache TTL settings
    3. Check for cache key pattern changes
    4. Monitor database query performance

    {{#is_warning}}
    ${var.slack_channel}
    {{/is_warning}}

    {{#is_alert}}
    ${var.critical_slack_channel}
    {{/is_alert}}
  EOT

  query = "sum(last_10m):sum:api.users.get.cache_miss{*} > 100"

  monitor_thresholds {
    critical = 100
    warning  = 50
  }

  notify_no_data      = false
  renotify_interval   = 0
  notify_audit        = false
  timeout_h           = 0
  include_tags        = true
  require_full_window = false  # 部分的なデータでも評価

  tags = [
    "service:${var.service_name}",
    "env:${var.environment}",
    "component:cache",
    "severity:medium",
    "type:performance"
  ]
}
```

**モニターID**: 10529857

**注意**: `cache_hit`メトリクスが生成されていないため、キャッシュヒット率ではなく、キャッシュミス数で監視しています。

#### 3. response_time.tf - API応答時間

```hcl
resource "datadog_monitor" "api_response_time" {
  name    = "[${var.service_name}] Slow API Response Time"
  type    = "metric alert"
  message = <<-EOT
    ## Slow API Response Time

    **p95 Response Time**: {{value}} seconds
    **Threshold**: {{threshold}} seconds
    **Environment**: ${var.environment}

    ### User Impact
    - Degraded user experience
    - Potential timeout errors
    - SEO impact

    ### Investigation Checklist
    1. Database query performance
    2. Cache hit rate
    3. CPU/Memory usage
    4. Active database connections
    5. Recent code deployments

    {{#is_warning}}
    ${var.slack_channel}
    {{/is_warning}}

    {{#is_alert}}
    ${var.critical_slack_channel}
    {{/is_alert}}
  EOT

  query = "avg(last_5m):avg:api.users.list.duration.95percentile{*} > 1.0"

  monitor_thresholds {
    critical = 1.0
    warning  = 0.5
  }

  notify_no_data      = false
  renotify_interval   = 0
  notify_audit        = false
  timeout_h           = 0
  include_tags        = true
  require_full_window = false  # 部分的なデータでも評価

  tags = [
    "service:${var.service_name}",
    "env:${var.environment}",
    "component:api",
    "severity:medium",
    "type:performance"
  ]
}
```

**モニターID**: 10529858

#### 4. service_down.tf - サービスダウン検知

```hcl
resource "datadog_monitor" "service_down" {
  name    = "[${var.service_name}] Service Down - No Health Check Response"
  type    = "metric alert"
  message = <<-EOT
    ## Service Down - No Health Check Response

    **Status**: Service appears to be down
    **Environment**: ${var.environment}
    **Last Check**: {{last_triggered_at}}

    ### Critical Actions Required
    1. Check application status
    2. Check application logs
    3. Restart service if needed
    4. Verify all dependencies (MySQL, Redis, Datadog Agent)

    ${var.critical_slack_channel} ${var.pagerduty_service}

    {{#is_recovery}}
    Service has recovered and is responding to health checks.
    {{/is_recovery}}
  EOT

  query = "sum(last_1m):sum:api.health.check{*} <= 0"

  monitor_thresholds {
    critical = 0
  }

  notify_no_data      = true
  no_data_timeframe   = 2
  renotify_interval   = 5
  notify_audit        = false
  timeout_h           = 0
  include_tags        = true
  require_full_window = false  # 部分的なデータでも評価

  tags = [
    "service:${var.service_name}",
    "env:${var.environment}",
    "severity:critical",
    "type:availability"
  ]
}
```

**モニターID**: 10529856

### 適用方法

```bash
cd terraform/monitors

# 初期化
terraform init

# 変更内容を確認
terraform plan

# 適用
terraform apply

# または自動承認
terraform apply -auto-approve
```

### トラブルシューティング

#### エラー: "The value provided for parameter 'query' is invalid"

**原因**: クエリ文字列に比較演算子と閾値が含まれていない

**解決方法**: クエリに`> 10`や`<= 0`などの比較演算子を追加

```hcl
# 間違い
query = "avg(last_5m):avg:api.users.create{*}"

# 正しい
query = "avg(last_5m):avg:api.users.create{*} > 10"
```

#### エラー: Monitor Status が "NO DATA"

**症状**: モニターが作成されているが、ステータスが「NO DATA」のまま変わらない

**原因**:
1. `require_full_window = true`（デフォルト）により、完全なデータウィンドウが必要
2. メトリクスは送信されているが、評価期間全体のデータが揃っていない
3. メトリクスの送信頻度が低い、または不定期

**解決方法**:

1. **`require_full_window = false`に設定**（推奨）:
```hcl
resource "datadog_monitor" "example" {
  # ...
  require_full_window = false  # この行を追加
}
```

2. **メトリクスを継続的に生成**:
```bash
# APIを定期的に叩いてメトリクスを生成
for i in {1..10}; do
  curl http://localhost:8080/health
  curl http://localhost:8080/api/users
  sleep 5
done
```

3. **Metrics Explorerで確認**:
   - [Metrics Explorer](https://ap1.datadoghq.com/metric/explorer)でメトリクスが存在するか確認
   - データポイントが時系列で連続しているか確認
   - 評価期間（`last_5m`など）に十分なデータポイントがあるか確認

4. **モニターの評価を待つ**:
   - モニターは1分ごとに評価されるため、2-3分待つ
   - 設定変更後、Terraformで`apply`して数分待つ

**確認コマンド**:
```bash
# terraform apply後、モニターを更新
cd terraform/monitors
terraform apply -auto-approve

# メトリクスを生成
for i in {1..10}; do
  curl -s http://localhost:8080/health > /dev/null
  sleep 6
done

# 数分待ってDatadog UIで確認
```

#### エラー: 403 Forbidden

**原因**: APIエンドポイントがリージョンと一致していない

**解決方法**: `main.tf`でリージョンに応じたAPIエンドポイントを設定

```hcl
# US1リージョン（デフォルト）
api_url = "https://api.datadoghq.com"

# AP1リージョン（日本）
api_url = "https://api.ap1.datadoghq.com"

# EU1リージョン
api_url = "https://api.datadoghq.eu"
```

### モニターの確認

作成されたモニターはDatadog UIで確認できます：

https://app.datadoghq.com/monitors/manage

タグで絞り込み：
- `service:datadog-tour-api`
- `env:production`
- `severity:high` / `severity:medium` / `severity:critical`

---

## まとめ

### 推奨アラート一覧

このプロジェクトで設定すべきアラート：

1. **ユーザー作成エラー率**: > 5% (Warning), > 10% (Critical)
2. **API応答時間**: > 500ms (Warning), > 1000ms (Critical)
3. **キャッシュヒット率**: < 80% (Warning), < 60% (Critical)
4. **異常なユーザー作成数**: 通常の3倍を超える
5. **サービスダウン**: ヘルスチェック0回（1分間）
6. **ユーザー総数減少**: 1時間で10%以上減少

### アラート設定のポイント

1. **適切な閾値**: 過去データを参考に設定
2. **明確な通知**: 問題、影響、対処手順を含める
3. **優先度の区別**: Critical、Warning、Infoを使い分け
4. **自動化**: Terraformでコード管理
5. **定期レビュー**: 月次でアラート設定を見直し

### 次のステップ

1. Datadog UIで推奨アラートを設定
2. Slackチャネルを作成して通知先を設定
3. Runbookを作成して対処手順を明確化
4. Terraformでアラート設定を自動化
5. 定期的にアラートをテストして有効性を確認

---

## 参考資料

- [Datadog Monitors Documentation](https://docs.datadoghq.com/monitors/)
- [Alerting Best Practices](https://docs.datadoghq.com/monitors/guide/best-practices/)
- [Terraform Datadog Provider](https://registry.terraform.io/providers/DataDog/datadog/latest/docs)
- [本プロジェクトのCustom Metricsガイド](./custom-metrics-guide.md)
