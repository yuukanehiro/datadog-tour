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

    ### Links
    - [View Cache Metrics](https://app.datadoghq.com/dashboard/cache-metrics)
    - [View Redis Logs](https://app.datadoghq.com/logs?query=source:redis)

    {{#is_warning}}
    ${var.slack_channel}
    {{/is_warning}}

    {{#is_alert}}
    ${var.critical_slack_channel}
    {{/is_alert}}

    {{#is_recovery}}
    Cache miss count has returned to normal levels.
    {{/is_recovery}}
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
  require_full_window = false

  tags = [
    "service:${var.service_name}",
    "env:${var.environment}",
    "component:cache",
    "severity:medium",
    "type:performance"
  ]
}
