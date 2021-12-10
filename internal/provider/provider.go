package provider

import (
	"context"
	"log"

	"github.com/ajbosco/segment-config-go/segment"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// Provider -
func New() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"access_token": {
				Description: "A Segment Config API token, more details about it can be found the [Config API documentation](https://segment.com/docs/config-api/authentication/).",
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("SEGMENT_ACCESS_TOKEN", nil),
			},
			"workspace": {
				Description: "The Segment workspace slug.",
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("SEGMENT_WORKSPACE", nil),
			},
			"unsupported_destination_config_props": {
				Description: "An array of destination configuration properties which are not supported by the Segment Config API and will result in an error when applying the plan.\n" +
					"Properties defined here get removed from the destination configuration before calling the API. These properties will need to be defined through the UI instead." +
					"Configuration properties of type `select` are the ones resulting in an error.",
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				DefaultFunc: func() (interface{}, error) { return []interface{}{}, nil },
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"segment_tracking_plan":      resourceTrackingPlan(),
			"segment_source":             resourceSegmentSource(),
			"segment_destination":        resourceSegmentDestination(),
			"segment_destination_filter": resourceSegmentDestinationFilter(),
		},
		DataSourcesMap: map[string]*schema.Resource{
			"segment_event_library": dataSourceEventLibrary(),
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
