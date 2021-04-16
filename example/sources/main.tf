
resource "source" "test_source" {
  provider     = segment
  catalog_name = "catalog/sources/javascript"
  source_name  = "test_source"

  config {
    allow_unplanned_track_events           = true
    allow_unplanned_identify_traits        = false
    allow_unplanned_group_traits           = true
    forwarding_blocked_events_to           = ""
    allow_unplanned_track_event_properties = false
    allow_track_event_on_violations        = true
    allow_identify_traits_on_violations    = true
    allow_group_traits_on_violations       = true
    forwarding_violations_to               = ""
    allow_track_properties_on_violations   = true
    common_track_event_on_violations       = "ALLOW"
    common_identify_event_on_violations    = "OMIT_TRAITS"
    common_group_event_on_violations       = "BLOCK"
  }
}