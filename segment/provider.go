package segment

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/uswitch/segment-config-go/segment"
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
		},
		ResourcesMap: map[string]*schema.Resource{
			"tracking_plan": resourceTrackingPlan(),
		},
		DataSourcesMap: map[string]*schema.Resource{
			"event_library": dataSourceEventLibrary(),
		},
		ConfigureContextFunc: providerConfigure,
	}
}

func providerConfigure(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
	var diags diag.Diagnostics

	accessToken := d.Get("access_token").(string)
	workSpace := d.Get("workspace").(string)

	if accessToken != "" && workSpace != "" {
		c := segment.NewClient(accessToken, workSpace)
		return c, diags
	}

	diags = append(diags, diag.Diagnostic{
		Severity: diag.Error,
		Summary:  "Unable to create Segment Config API client",
		Detail:   "Access token and workspace values cannot be empty",
	})
	return nil, diags
}
