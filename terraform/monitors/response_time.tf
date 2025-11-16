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
    2. Cache hit rate: [View Dashboard](https://app.datadoghq.com/dashboard/cache-metrics)
    3. CPU/Memory usage
    4. Active database connections
    5. Recent code deployments

    ### APM Traces
    [View Slow Traces](https://app.datadoghq.com/apm/traces?query=service:${var.service_name}%20resource_name:"GET%20/api/users"%20@duration:>500ms)

    ### Common Causes
    - N+1 query problem
    - Missing database indexes
    - Cache invalidation
    - High database load

    {{#is_warning}}
    ${var.slack_channel}
    {{/is_warning}}

    {{#is_alert}}
    ${var.critical_slack_channel}
    {{/is_alert}}

    {{#is_recovery}}
    Response time has returned to normal levels.
    {{/is_recovery}}
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
  require_full_window = false

  tags = [
    "service:${var.service_name}",
    "env:${var.environment}",
    "component:api",
    "severity:medium",
    "type:performance"
  ]
}
