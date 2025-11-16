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

    ### Commands
    ```bash
    # Check status
    docker-compose ps api

    # Check logs
    docker-compose logs api --tail 50

    # Restart service
    make restart-api
    ```

    ### Escalation
    If not resolved in 5 minutes, escalate immediately.

    ${var.critical_slack_channel} ${var.pagerduty_service}

    {{#is_recovery}}
    Service has recovered and is responding to health checks.
    {{/is_recovery}}
  EOT

  query = "sum(last_1m):sum:api.health.check{*} <= 0"

  monitor_thresholds {
    critical = 0
  }

  notify_no_data    = true
  no_data_timeframe = 2
  renotify_interval = 5
  notify_audit      = false
  timeout_h         = 0
  include_tags      = true
  require_full_window = false

  tags = [
    "service:${var.service_name}",
    "env:${var.environment}",
    "severity:critical",
    "type:availability"
  ]
}
