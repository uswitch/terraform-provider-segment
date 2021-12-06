provider "segment" {
  access_token = "test_access_token"       # or via SEGMENT_ACCESS_TOKEN env var.
  workspace    = "test_workspace"          # or via SEGMENT_WORKSPACE env var.
  unsupported_destination_config_props = [ # Optional
    "appboy/datacenter"
  ]
}
