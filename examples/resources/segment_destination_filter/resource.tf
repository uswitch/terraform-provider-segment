# Filter to drop events
resource "segment_destination_filter" "drop_events" {
  destination = "test_source/test_destination"
  title       = "Checkout events only"
  description = "Only send checkout events to destination"
  condition   = "event != \"Checkout\"" # Invert condition to drop all events but `Checkout`
  enabled     = true
  actions {
    drop {}
  }
}

# Filter to sample data
resource "segment_destination_filter" "sample_events" {
  destination = "test_source/test_destination"
  title       = "Users accepting marketing"
  description = "Sends 10% of users accepting marketing permissions"
  condition   = "context.castPermissions.marketing = true"
  enabled     = true
  actions {
    sample {
      percent = 0.1
      path    = "anomnymousId"
    }
  }
}

# Filter to block fields
resource "segment_destination_filter" "checkout_block_ip_postcode" {
  destination = "test_source/test_destination"
  title       = "No IP and postcode on Checkout"
  description = "Prevents IP address and postcode being sent on checkout"
  condition   = "event == \"Checkout\""
  enabled     = true
  actions {
    block_fields {
      context    = ["client.ip"]
      properties = ["postcode"]
    }
  }
}

# Filter to allow fields
resource "segment_destination_filter" "checkout_allow_price_product" {
  destination = "test_source/test_destination"
  title       = "Checkout price and product only"
  description = "Allows only price and product fields to go through on a Checkout event"
  condition   = "event == \"Checkout\""
  enabled     = true
  actions {
    allow_fields {
      properties = ["price", "product"]
    }
  }
}
