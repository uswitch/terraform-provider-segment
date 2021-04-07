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
		CreateContext: resourceTrackingPlanCreate, // todo: update to merge json with event library
		ReadContext:   resourceTrackingPlanRead,
		UpdateContext: resourceTrackingPlanUpdate, // todo: update to merge json with event library
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
			"rules_json_file": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringIsJSON,
			},
			"import_from": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.StringIsJSON,
				},
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

	// read rules from json file
	var rules segment.RuleSet
	if rulesJSON, ok := d.GetOk("rules_json_file"); ok {
		err := json.Unmarshal([]byte(rulesJSON.(string)), &rules)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	// read from event library
	var eventLibs []EventLibrary
	if eventLibsIntfcs, ok := d.GetOk("import_from"); ok {
		var eventLibsList = eventLibsIntfcs.([]interface{})
		var err error
		if eventLibs, err = readEventLibs(eventLibsList); err != nil {
			return diag.FromErr(err)
		}
	}

	// flatten event libraries
	eventLibsFlat := flattenEventLibs(eventLibs)

	// Convert event library events to Segment type events
	mergedEvents := eventLibsFlat.convertToSegmentEvents()

	// Merge json schema events with ones from the event library.
	for _, ruleEvnt := range rules.Events {
		exists := false
		for i, mergedEvnt := range mergedEvents {
			if ruleEvnt.Name == mergedEvnt.Name {
				mergedEvents[i] = ruleEvnt
				exists = true
				break
			}
		}
		if !exists {
			mergedEvents = append(mergedEvents, ruleEvnt)
		}
	}
	rules.Events = mergedEvents

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

	// get a tracking plan by ID
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
	if err := d.Set("rules_json_file", string(rulesJSON)); err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func resourceTrackingPlanUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*segment.Client)

	tpID := d.Id()
	if d.HasChange("display_name") || d.HasChange("rules_json_file") {
		displayName := d.Get("display_name").(string)
		// unmarshal rules
		var rules segment.RuleSet
		json.Unmarshal([]byte(d.Get("rules_json_file").(string)), &rules)

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

func flattenEventLibs(eventLibs []EventLibrary) EventLibrary {
	var eventLibsFlat EventLibrary
	for _, evtLib := range eventLibs {
		eventLibsFlat.Events = append(eventLibsFlat.Events, evtLib.Events...)
	}
	return eventLibsFlat
}

func readEventLibs(eventLibsSlice []interface{}) ([]EventLibrary, error) {
	var eventLibs []EventLibrary
	for _, eventLibJSONIntfc := range eventLibsSlice {
		var eventLib EventLibrary
		if err := json.Unmarshal([]byte(eventLibJSONIntfc.(string)), &eventLib); err != nil {
			return nil, err
		}
		eventLibs = append(eventLibs, eventLib)
	}
	return eventLibs, nil
}
