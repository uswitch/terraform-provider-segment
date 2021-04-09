terraform {
  required_providers {
    segment = {
      source = "uswitch.com/segment/segment"
      version = ">= 0.1.0"
    }
  }
}

provider "segment" {
  access_token = "<access-token>"
	workspace = "<workspace>"
}

module "tracking_plans" {
    source = "./tracking_plans"
}
