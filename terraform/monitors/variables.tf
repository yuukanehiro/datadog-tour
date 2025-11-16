variable "datadog_api_key" {
  description = "Datadog API Key"
  type        = string
  sensitive   = true
}

variable "datadog_app_key" {
  description = "Datadog Application Key"
  type        = string
  sensitive   = true
}

variable "environment" {
  description = "Environment name (production, staging, development)"
  type        = string
  default     = "production"
}

variable "service_name" {
  description = "Service name"
  type        = string
  default     = "datadog-tour-api"
}

variable "slack_channel" {
  description = "Slack channel for alerts"
  type        = string
  default     = "@slack-alerts"
}

variable "critical_slack_channel" {
  description = "Slack channel for critical alerts"
  type        = string
  default     = "@slack-critical"
}

variable "pagerduty_service" {
  description = "PagerDuty service for oncall alerts"
  type        = string
  default     = "@pagerduty-oncall"
}
