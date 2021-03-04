terraform {
  required_providers {
    segment = {
      source = "uswitch.com/segment/segment"
      version = "~> 0.1.0"
    }
  }
}

provider "segment" {
  access_token = "pGXsHj5W6O7D0H_rNOuFQcJNNg9HkV8Puj6WTT2fEag.4fcV8s0zmYaA_MiJICKba0z6Jyzg-x-S2FQGxCxNv8M"
	workspace = "uswitch-sandbox"
}

resource "tracking_plan" "test" {
    provider = segment

    display_name = "AW test tracking plan"
    rules = file("./rules/test_tracking_plan/rules.json")
}