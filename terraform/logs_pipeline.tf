resource "datadog_logs_custom_pipeline" "error_notification_pipeline" {
  name       = "Error Notification Pipeline"
  is_enabled = true

  filter {
    query = "service:${var.service_name} level:error"
  }

  processor {
    category_processor {
      name       = "Categorize error notifications"
      is_enabled = true
      target     = "alert.type"

      category {
        name = "system_alert"

        filter {
          query = "@error.notify:true"
        }
      }

      category {
        name = "expected_error"

        filter {
          query = "@error.notify:false"
        }
      }
    }
  }

  processor {
    attribute_remapper {
      name                 = "Map error.type to error_category"
      is_enabled           = true
      source_type          = "attribute"
      sources              = ["error.type"]
      target               = "error_category"
      target_type          = "attribute"
      preserve_source      = true
      override_on_conflict = false
    }
  }
}
