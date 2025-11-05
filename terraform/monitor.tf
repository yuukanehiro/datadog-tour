resource "datadog_monitor" "system_error_alert" {
  name    = "[${var.environment}] System Error Alert - ${var.service_name}"
  type    = "log alert"
  message = <<-EOT
    {{#is_alert}}
    ⚠️ **System Error Detected**

    **Service:** ${var.service_name}
    **Environment:** ${var.environment}
    **Error Count:** {{value}} errors in last 5 minutes

    **Details:**
    - Alert Type: System Alert (error.notify:true)
    - Service: {{service.name}}
    - Error Type: {{error_category.name}}

    **Action Required:**
    Check [Datadog Logs Explorer](https://app.datadoghq.com/logs?query=service%3A${var.service_name}%20level%3Aerror%20error.notify%3Atrue) for details.

    ${var.slack_channel}
    {{/is_alert}}

    {{#is_recovery}}
    ✅ **Error Resolved**

    System errors have stopped in ${var.service_name}.
    {{/is_recovery}}
  EOT

  query = "logs(\"service:${var.service_name} status:error @error.notify:true\").index(\"*\").rollup(\"count\").last(\"5m\") > 0"

  monitor_thresholds {
    critical = 0
  }

  notify_no_data      = false
  renotify_interval   = 60
  timeout_h           = 0
  include_tags        = true
  require_full_window = false

  tags = [
    "service:${var.service_name}",
    "env:${var.environment}",
    "team:backend",
    "alert_type:system_error"
  ]
}

resource "datadog_monitor" "error_rate_spike" {
  name    = "[${var.environment}] High Error Rate - ${var.service_name}"
  type    = "log alert"
  message = <<-EOT
    {{#is_alert}}
    🚨 **High Error Rate Detected**

    **Service:** ${var.service_name}
    **Environment:** ${var.environment}
    **Error Count:** {{value}} errors in last 15 minutes

    **Threshold:** More than 10 system errors detected

    **Action Required:**
    Investigate the spike in errors immediately.
    Check [Datadog Logs Explorer](https://app.datadoghq.com/logs?query=service%3A${var.service_name}%20level%3Aerror%20error.notify%3Atrue) for details.

    ${var.slack_channel}
    {{/is_alert}}

    {{#is_warning}}
    ⚠️ **Warning: Elevated Error Rate**

    Error rate is elevated but below critical threshold.
    {{/is_warning}}

    {{#is_recovery}}
    ✅ **Error Rate Normalized**

    Error rate has returned to normal levels.
    {{/is_recovery}}
  EOT

  query = "logs(\"service:${var.service_name} status:error @error.notify:true\").index(\"*\").rollup(\"count\").last(\"15m\") > 10"

  monitor_thresholds {
    critical = 10
    warning  = 5
  }

  notify_no_data      = false
  renotify_interval   = 120
  timeout_h           = 0
  include_tags        = true
  require_full_window = false

  tags = [
    "service:${var.service_name}",
    "env:${var.environment}",
    "team:backend",
    "alert_type:error_rate"
  ]
}
