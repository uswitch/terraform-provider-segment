terraform {
  required_providers {
    segment = {
      versions = ["0.0.9"]
      source = "uswitch.com/segment/segment"
    }
  }
}

provider "segment" {}

resource "tracking_plan" "test" {
    provider = segment
    
    display_name = "DP test tracking plan update2"
    rules = ""
    
}
