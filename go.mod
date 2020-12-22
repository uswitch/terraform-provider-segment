module github.com/uswitch/terraform-provider-segment

go 1.15

require github.com/hashicorp/terraform-plugin-sdk/v2 v2.2.0
require github.com/uswitch/segment-config-go v0.0.0
replace github.com/uswitch/segment-config-go v0.0.0 => "../segment-config-go"
