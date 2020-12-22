package segment

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
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
	  DataSourcesMap: map[string]*schema.Resource{},
	  ConfigureContextFunc: providerConfigure,
	//   ConfigKeys: map[string]*schema.Schema{}, Need to pass the segment api key?
	}
  }


  func providerConfigure(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {

	  var diags diag.Diagnostics
	//   accessToken := d.Get("accessToken").(string)
	  accessToken := "t5dwvaonRp-UqFZJX0iOStP1Cj8_Ea25z8FuxDQWL8U.sHdro1niwYG3oqrnVxSgBUEYNdNPUNVvNiRRUutCHbc"
	//   workSpace := d.Get("workSpace").(string)
	  workSpace := "uswitch-sandbox"

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

  
  