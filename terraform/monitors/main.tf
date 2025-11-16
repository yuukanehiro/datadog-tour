terraform {
  required_version = ">= 1.0"

  required_providers {
    datadog = {
      source  = "DataDog/datadog"
      version = "~> 3.0"
    }
  }
}

provider "datadog" {
  api_key = var.datadog_api_key
  app_key = var.datadog_app_key

  # Datadog API endpoint - adjust based on your region
  # US1 (default): https://api.datadoghq.com
  # AP1 (Japan):   https://api.ap1.datadoghq.com
  # EU1:           https://api.datadoghq.eu
  api_url = "https://api.ap1.datadoghq.com"
}
