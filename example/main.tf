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
	workspace = "<workspace>"
}

data "event_library" "generic_events" {
  event {
    name = "generic event 1"
    description = "first generic event"

    property {
      type = ""
      name = ""
      required = true
    }

    property {
      type = ""
      name = ""
      required = false
    }
  }

  event {
    name = "generic event 2"
    description = "second generic event"

    property {
      type = ""
      name = ""
      required = true
    }

    property {
      type = ""
      name = ""
      required = false
    }
  }
}

resource "tracking_plan" "test" {
  provider = segment

  display_name = "AW test tracking plan"
  rules {
    global {
      property {
        type = ""
        name = ""
        required = true
      }
    }

    event {
      name = "specific event 1"
      description = "first specific event"

      property {
        type = ""
        name = ""
        required = true
      }

      property {
        type = ""
        name = ""
        required = false
      }
    }

    event {
      name = "specific event 2"
      description = "second specific event"

      property {
        type = ""
        name = ""
        required = true
      }

      property {
        type = ""
        name = ""
        required = false
      }
    }

    identify {
      property {
        type = ""
        name = ""
        required = true
      }
    }

    group {
      property {
        type = ""
        name = ""
        required = true
      }
    }
  }


  rules_source_file = file("./rules/test_tracking_plan/rules.json")  # overrides everything
  import_from = [data.event_library.generic_events]  # merges together with already defined data either via file or code
}
