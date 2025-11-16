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

    ### Runbook
    See: https://wiki.example.com/runbooks/user-creation-errors

    {{#is_warning}}
    ${var.slack_channel}
    {{/is_warning}}

    {{#is_alert}}
    ${var.critical_slack_channel} ${var.pagerduty_service}
    {{/is_alert}}

    {{#is_recovery}}
    The error rate has returned to normal levels.
    {{/is_recovery}}
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
  require_full_window = false

  tags = [
    "service:${var.service_name}",
    "env:${var.environment}",
    "team:backend",
    "severity:high",
    "type:error_rate"
  ]
}
