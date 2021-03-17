terraform {
  required_providers {
    segment = {
      source = "uswitch.com/segment/segment"
      version = "~> 0.1.0"
    }
  }
}

provider "segment" {
  access_token = "<access-token>"
	workspace = "uswitch-sandbox"
}

resource "tracking_plan" "test" {
    provider = segment

    display_name = "AW test tracking plan"
    rules = file("./rules/test_tracking_plan/rules.json")
}