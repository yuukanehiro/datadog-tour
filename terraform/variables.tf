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

variable "datadog_api_url" {
  description = "Datadog API URL"
  type        = string
  default     = "https://api.datadoghq.com"
}

variable "service_name" {
  description = "Service name to monitor"
  type        = string
  default     = "datadog-tour-api"
}

variable "environment" {
  description = "Environment name"
  type        = string
  default     = "development"
}

variable "slack_channel" {
  description = "Slack channel for notifications (e.g., @slack-alerts)"
  type        = string
}

variable "slack_webhook_url" {
  description = "Slack Webhook URL for notifications"
  type        = string
  sensitive   = true
}
