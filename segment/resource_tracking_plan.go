package segment

import (
	"context"
	"regexp"

	// "encoding/json"
	// "fmt"
	"log"

	//   "strconv"
	//   "time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/uswitch/segment-config-go/segment"
)

var client segment.Client

func resourceTrackingPlan() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceTrackingPlanCreate,
		ReadContext:   resourceTrackingPlanRead,
		UpdateContext: resourceTrackingPlanUpdate,
		DeleteContext: resourceTrackingPlanDelete,
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"display_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"rules": { //TODO: The rules schema needs more work. an example is provided in the resourceTrackingPlan.txt
				Type:        schema.TypeString,
				Optional:    true,
				Description: "List of identify traits, group traits and events",
				ValidateFunc: validation.StringIsJSON,
			},
			"create_time": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"update_time": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
		},
	}
}

func resourceTrackingPlanCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	var diags diag.Diagnostics
	// fmt.Println(reflect.TypeOf(m))
	client := m.(*segment.Client)
	
	log.Println("[INFO] client")
	log.Println(client)
	tp := segment.TrackingPlan{

		DisplayName: d.Get("display_name").(string),
		// Rules:       d.Get("rules").(segment.Rules),
		Rules: d.Get("rules").(string),
	}
	respText, err := client.CreateTrackingPlan(tp)
	if err != nil {
		return diag.FromErr(err)
	}

	// setId shoud utilise the calculated name part in the schema
	re := regexp.MustCompile(`rs_.*$`)
	trackingPlanID := re.FindString(respText.Name)
	d.SetId(trackingPlanID)

	return diags
	// TODO: return resourceTrackingPlanRead(ctx, d, m)
}

func resourceTrackingPlanUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	var diags diag.Diagnostics
	client := m.(*segment.Client)

	tpID := d.Id() 
	if d.HasChange("display_name") || d.HasChange("rules") {
		displayName := d.Get("display_name").(string)
		rules := d.Get("rules").(string)

		data := segment.TrackingPlan{
			DisplayName: displayName,
			Rules:       rules,
		}
		tp, err := client.UpdateTrackingPlan(tpID, data)

		log.Println(tp)

		if err != nil {
			return diag.FromErr(err)
		}

		return resourceTrackingPlanRead(ctx, d, m)
	}

	//TODO: invoke read to update the state return resourceTrackingPlanRead(ctx, d, m)

	return diags
}

func resourceTrackingPlanRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	// 	client := m.(*segment.Client)

	// 	// send a get request to get a tracking plan
	// 	tp, err := client.GetTrackingPlan(d.Id())
	// 	if err != nil {
	// 		return diag.FromErr(err)
	// 	}

	// 	// set the variables
	// 	if err := d.Set("display_name", tp.DisplayName); err != nil {
	// 		return diag.FromErr(err)
	// 	}
	// 	if err := d.Set("rules", tp.Rules); err != nil {
	// 		return diag.FromErr(err)
	// 	}
	// 	if err := d.Set("create_time", tp.CreateTime); err != nil {
	// 		return diag.FromErr(err)
	// 	}
	// 	if err := d.Set("update_time", tp.UpdateTime); err != nil {
	// 		return diag.FromErr(err)
	// 	}
	return diags
}

func resourceTrackingPlanDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	log.Println("resourceTrackingplan Delete")
	client := m.(*segment.Client)
	id := d.Id()
	log.Println(id)
	err := client.DeleteTrackingPlan(d.Id())
		if err != nil {
			return diag.FromErr(err)
		}

		d.SetId("")

	return diags
}
