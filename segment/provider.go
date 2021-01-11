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
	// 		  Required:    true,
	// 		  DefaultFunc: schema.EnvDefaultFunc("ACCESS_TOKEN", nil),
	// 		},
	// 		"workSpace": &schema.Schema{
	// 		  Type: 		schema.TypeString,
	// 		  Required:  	true, 
	// 		  DefaultFunc:  schema.EnvDefaultFunc("SEGMENT_WORKSPACE", nil),
	// 		},
	// 	},
	  ResourcesMap: map[string]*schema.Resource{
		  "tracking_plan": resourceTrackingPlan(),  
	  },
	  DataSourcesMap: map[string]*schema.Resource{},
	  ConfigureContextFunc: providerConfigure,
	}
  }


  func providerConfigure(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {

	  var diags diag.Diagnostics
	//   accessToken := d.Get("accessToken").(string)
	  accessToken := "pGXsHj5W6O7D0H_rNOuFQcJNNg9HkV8Puj6WTT2fEag.4fcV8s0zmYaA_MiJICKba0z6Jyzg-x-S2FQGxCxNv8M"
	//   TODO: use env variable for authorization
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

  
//   func validateAccessToken(v interface{}, k string) (diag.Diagnostics) {
// 	  if v == nil || v.(string) == "" {
// 		  return nil
// 	  }
// 	  return nil
//   }
