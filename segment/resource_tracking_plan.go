package segment

import (
	"context"
	"encoding/json"
	"regexp"

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
			"rules": {
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "List of identify traits, group traits and events",
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
	client := m.(*segment.Client)

	// unmarshal rules
	var rules segment.Rules
	json.Unmarshal([]byte(d.Get("rules").(string)), &rules)

	// construct the tracking plan
	tp := segment.TrackingPlan{
		DisplayName: d.Get("display_name").(string),
		Rules:       rules,
	}
	response, err := client.CreateTrackingPlan(tp)
	if err != nil {
		return diag.FromErr(err)
	}

	// setId shoud utilise the calculated name part in the schema
	re := regexp.MustCompile(`rs_.*$`)
	trackingPlanID := re.FindString(response.Name)
	d.SetId(trackingPlanID)

	return resourceTrackingPlanRead(ctx, d, m)
}

func resourceTrackingPlanRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*segment.Client)

	// get a tracking plan by name
	tp, err := client.GetTrackingPlan(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	// set the variables
	if err := d.Set("name", tp.Name); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("display_name", tp.DisplayName); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("create_time", tp.CreateTime.String()); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("update_time", tp.UpdateTime.String()); err != nil {
		return diag.FromErr(err)
	}

	// convert Rules to JSON
	rulesJSON, err := json.MarshalIndent(tp.Rules, "", "  ")
	if err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("rules", string(rulesJSON)); err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func resourceTrackingPlanUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*segment.Client)

	tpID := d.Id()
	if d.HasChange("display_name") || d.HasChange("rules") {
		displayName := d.Get("display_name").(string)
		// unmarshal rules
		var rules segment.Rules
		json.Unmarshal([]byte(d.Get("rules").(string)), &rules)

		data := segment.TrackingPlan{
			DisplayName: displayName,
			Rules:       rules,
		}
		_, err := client.UpdateTrackingPlan(tpID, data)

		if err != nil {
			return diag.FromErr(err)
		}

		return resourceTrackingPlanRead(ctx, d, m)
	}
	return diags
}

func resourceTrackingPlanDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*segment.Client)

	err := client.DeleteTrackingPlan(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")

	return resourceTrackingPlanRead(ctx, d, m)
}
