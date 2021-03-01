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
			"access_token": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("SEGMENT_ACCESS_TOKEN", nil),
			},
			"workspace": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("SEGMENT_WORKSPACE", nil),
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"tracking_plan": resourceTrackingPlan(),
		},
		DataSourcesMap:       map[string]*schema.Resource{},
		ConfigureContextFunc: providerConfigure,
	}
}

func providerConfigure(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
	var diags diag.Diagnostics

	accessToken := d.Get("access_token").(string)
	workSpace := d.Get("workspace").(string)
	if accessToken != "" && workSpace != "" {
		c, err := segment.NewClient(&accessToken, &workSpace)
		if err != nil {
			return nil, diag.FromErr(err)
		}

		return c, diags
	}

	c, err := segment.NewClient(nil, nil)
	if err != nil {
		return nil, diag.FromErr(err)
	}
	return &c, nil
}
