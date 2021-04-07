data "event_library" "form_events" {
  provider = segment

  event {
    name = "Form Submitted"
    description = "fired when a user submits a form."

    property {
      type = ["string"]
      name = "form_id"
      description = "the ID of the form"
      required = true
    }

    property {
      type = ["string", "null"]
      name = "name"
      description = "the name of the form"
      required = false
    }

    property {
      type = ["integer"]
      name = "form_orientation"
      description = "orientation of the form"
      required = false
    }
  }
}

output "form_events" {
    value = data.event_library.form_events.json
}