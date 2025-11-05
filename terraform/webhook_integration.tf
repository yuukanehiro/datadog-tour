resource "datadog_webhook" "slack_test" {
  name = "slack-test"
  url  = var.slack_webhook_url

  # Use Slack message format
  encode_as = "json"

  # Custom payload for Slack
  payload = jsonencode({
    text = "$EVENT_TITLE"
    attachments = [
      {
        color = "$ALERT_TRANSITION"
        title = "$EVENT_TITLE"
        text  = "$TEXT_ONLY_MSG"
        fields = [
          {
            title = "Priority"
            value = "$PRIORITY"
            short = true
          },
          {
            title = "Alert ID"
            value = "$ID"
            short = true
          },
          {
            title = "Link"
            value = "$LINK"
            short = false
          }
        ]
        footer      = "Datadog Monitor Alert"
        footer_icon = "https://datadog-docs.imgix.net/images/dd_icon_32x32.png"
      }
    ]
  })

  custom_headers = jsonencode({
    "Content-Type" = "application/json"
  })
}
