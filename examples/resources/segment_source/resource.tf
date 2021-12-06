# Simple source
resource "segment_source" "simple_test" {
  catalog_name = "catalog/sources/ruby"
  source_name  = "simple test"
}

# Source with tracking plan
resource "segment_source" "test_with_tracking_plan" {
  catalog_name = "catalog/source/ruby"
  source_name  = "test with tracking plan"

  tracking_plan = "tr_123_abc" # Segment ID of the tracking plan
}

# Source with tracking plan and custom schema config
resource "segment_source" "test_with_custom_config" {
  catalog_name = "catalog/source/ruby"
  source_name  = "test with custom config"

  tracking_plan = "tr_123_abc" # Segment ID of the tracking plan
  schema_config {
    allow_group_traits_on_violations       = true
    allow_identify_traits_on_violations    = true
    allow_track_event_on_violations        = true
    allow_track_properties_on_violations   = false
    allow_unplanned_group_traits           = true
    allow_unplanned_identify_traits        = true
    allow_unplanned_track_event_properties = true
    allow_unplanned_track_events           = true
    common_group_event_on_violations       = "ALLOW"
    common_identify_event_on_violations    = "OMIT_TRAITS"
    common_track_event_on_violations       = "ALLOW"
    forward_violations_to                  = "segment_violations" # Source needs to be defined first
    forward_blocked_events_to              = "segment_blocks"     # Source needs to be defined first
  }
}
