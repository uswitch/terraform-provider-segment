resource "tracking_plan" "test" {
  provider = segment

  display_name = "AW test tracking plan"

  # Import tracking plan specific rules in JSON schema format
  rules_json_file = file("${path.module}/rules.json")

  # Import event libraries
  import_from = [
    file("${path.module}/../event_libs/page_events.json"),
    file("${path.module}/../event_libs/form_events.json"),
  ]
}