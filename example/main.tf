terraform {
  required_providers {
    segment = {
      source  = "uswitch.com/segment/segment"
      version = "0.2.0"
    }
  }
}

provider "segment" {
  access_token = "<access-token>"
  workspace    = "uswitch-sandbox"
}

module "tracking_plans" {
  source = "./tracking_plans"
}

module "sources" {
  source = "./sources"
}

