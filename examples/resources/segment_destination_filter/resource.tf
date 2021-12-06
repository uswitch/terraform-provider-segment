# Filter to drop events
resource "segment_destination_filter" "drop_events" {
  destination = "test_destination"
  title       = "Checkout events only"
  description = "Only send checkout events to destination"
  condition   = "event != \"Checkout\"" # Invert condition to drop all events but `Checkout`
  enabled     = true
  actions {
    drop {}
  }
}

# Filter to sample data

# Filter to blacklist fields

# Filter to whitelist fields
