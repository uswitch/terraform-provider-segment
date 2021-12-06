# Simple tracking plan
resource "segment_tracking_plan" "simple" {
  display_name    = "Simple Tracking Plan"
  rules_json_file = file("plans/simple/rules.json")
}

# Tracking plan with an event library attached
data "segment_event_library" "common" {
  json_file = file("plans/libs/common/rules.json")
}

resource "segment_tracking_plan" "complex" {
  display_name    = "Complex Tracking Plan"
  rules_json_file = file("plans/simple/rules.json")

  import_from = jsonencode([
    jsondecode(data.segment_event_library.common.json)
  ])
}
