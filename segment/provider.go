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
		//   Schema: map[string]*schema.Schema{
		// 		"accessToken": &schema.Schema{
		// 		  Type:        schema.TypeString,
		// 		  Optional:    true,
		// 		//   DefaultFunc: schema.EnvDefaultFunc("segmentAccessToken", nil),
		// 		  Default: "",
		// 		},
		// 	},
		ResourcesMap: map[string]*schema.Resource{
			"tracking_plan": resourceTrackingPlan(),
		},
		DataSourcesMap:       map[string]*schema.Resource{},
		ConfigureContextFunc: providerConfigure,
		//   ConfigKeys: map[string]*schema.Schema{}, Need to pass the segment api key?
	}
}

func providerConfigure(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {

	var diags diag.Diagnostics
	accessToken := d.Get("accessToken").(string)
	workSpace := d.Get("workSpace").(string)

	if (accessToken != "") && (workSpace != "") {
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
