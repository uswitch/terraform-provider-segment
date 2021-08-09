package segment

import (
	"context"
	"log"

	"github.com/ajbosco/segment-config-go/segment"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// Provider -
func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"access_token": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("SEGMENT_ACCESS_TOKEN", nil),
			},
			"workspace": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("SEGMENT_WORKSPACE", nil),
			},
			"unsupported_destination_config_props": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				DefaultFunc: func() (interface{}, error) { return []interface{}{}, nil },
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"tracking_plan":              resourceTrackingPlan(),
			"source":                     resourceSegmentSource(),
			"destination":                resourceSegmentDestination(),
			"segment_destination_filter": resourceSegmentDestinationFilter(),
		},
		DataSourcesMap: map[string]*schema.Resource{
			"event_library": dataSourceEventLibrary(),
		},
		ConfigureContextFunc: providerConfigure,
	}
}

func providerConfigure(_ context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
	var diags diag.Diagnostics

	accessToken := d.Get("access_token").(string)
	workSpace := d.Get("workspace").(string)

	if accessToken != "" && workSpace != "" {
		c := segment.NewClient(accessToken, workSpace)
		if c != nil {
			return ProviderMetadata{
				Client:                           *c,
				Workspace:                        workSpace,
				IsDestinationConfigPropSupported: isDestinationConfigPropSupported(d),
			}, diags
		}
	}

	diags = append(diags, diag.Diagnostic{
		Severity: diag.Error,
		Summary:  "Unable to create Segment Config API client",
		Detail:   "Access token and workspace values cannot be empty",
	})
	return nil, diags
}

// Provides a way to skip some destination configuration properties when sending them through the Config API.
// This is a workaround to circumvent Segment's bugs in the API
func isDestinationConfigPropSupported(d *schema.ResourceData) func(destination string, key string) bool {
	return func(destination string, key string) bool {
		exclusions := d.Get("unsupported_destination_config_props").(*schema.Set)
		excluded := exclusions.Contains(destination+"/"+key) || exclusions.Contains(key)

		if excluded {
			log.Printf("Excluding config %s/%s", destination, key)
		}

		return !excluded
	}
}

type ProviderMetadata struct {
	Client                           segment.Client
	Workspace                        string
	IsDestinationConfigPropSupported func(destination string, key string) bool
}
