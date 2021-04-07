data "event_library" "page_events" {
  provider = segment

  event {
    name = "Page Viewed"
    description = "fired when a user views a page."

    property {
      type = ["string"]
      name = "url"
      required = true
    }

    property {
      type = ["string", "null"]
      name = "title"
      description = "the title of the page"
      required = false
    }
  }
}

output "page_events" {
    value = data.event_library.page_events.json
}