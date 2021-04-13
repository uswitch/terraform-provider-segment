package segment

import (
	"context"
	"encoding/json"
	"regexp"

	"github.com/ajbosco/segment-config-go/segment"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceTrackingPlan() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceTrackingPlanCreate,
		ReadContext:   resourceTrackingPlanRead,
		UpdateContext: resourceTrackingPlanUpdate,
		DeleteContext: resourceTrackingPlanDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
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
				Type:             schema.TypeString,
				Optional:         true,
				ValidateFunc:     validation.StringIsJSON,
				DiffSuppressFunc: diffRulesJSONState,
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

func removeRulesJSONWhitespace(input string) string {
	var decodedStr segment.RuleSet
	if err := json.Unmarshal([]byte(input), &decodedStr); err != nil {
		panic(err)
	}

	encodedStr, err := json.Marshal(decodedStr)
	if err != nil {
		panic(err)
	}
	return string(encodedStr)
}

func diffRulesJSONState(k, old, new string, d *schema.ResourceData) bool {
	encodedNew := removeRulesJSONWhitespace(new)
	return old == encodedNew
}

func resourceTrackingPlanCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	meta := m.(SegmentMetadata)

	// Read tracking plan rules
	var tpRules segment.RuleSet
	if rulesJSON, ok := d.GetOk("rules_json_file"); ok {
		if err := json.Unmarshal([]byte(rulesJSON.(string)), &tpRules); err != nil {
			return diag.FromErr(err)
		}
	}

	// Read event library rules
	var eventLibs []segment.RuleSet
	if eventLibsIntfcs, ok := d.GetOk("import_from"); ok {
		var eventLibsJSONArray = eventLibsIntfcs.([]interface{})
		var err error
		if eventLibs, err = readEventLibs(eventLibsJSONArray); err != nil {
			return diag.FromErr(err)
		}
	}

	// Flatten event libraries
	eventLibsFlat := flattenEventLibs(eventLibs)

	// Merge json schema events with ones from the event library
	tpRules.Events = mergeEvents(eventLibsFlat.Events, tpRules.Events)

	// Construct the tracking plan
	tp := segment.TrackingPlan{
		DisplayName: d.Get("display_name").(string),
		Rules:       tpRules,
	}
	response, err := meta.client.CreateTrackingPlan(tp)
	if err != nil {
		return diag.FromErr(err)
	}

	// SetId shoud utilise the calculated name part in the schema
	re := regexp.MustCompile(`rs_.*$`)
	trackingPlanID := re.FindString(response.Name)
	d.SetId(trackingPlanID)

	return resourceTrackingPlanRead(ctx, d, m)
}

func resourceTrackingPlanRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	meta := m.(SegmentMetadata)

	tp, err := meta.client.GetTrackingPlan(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
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

	// Filter out library events so that we can accurately set/compare the source json file
	// Read event library rules
	var eventLibs []segment.RuleSet
	if eventLibsIntfcs, ok := d.GetOk("import_from"); ok {
		var eventLibsJSONArray = eventLibsIntfcs.([]interface{})
		var err error
		if eventLibs, err = readEventLibs(eventLibsJSONArray); err != nil {
			return diag.FromErr(err)
		}
	}

	// Flatten event libraries
	eventLibsFlat := flattenEventLibs(eventLibs)

	// Remove library events
	var sourceEvents []segment.Event
	for _, evnt := range tp.Rules.Events {
		inEventLib := false
		for _, eventLibEvt := range eventLibsFlat.Events {
			if evnt.Name == eventLibEvt.Name {
				inEventLib = true
				break
			}
		}
		if !inEventLib {
			sourceEvents = append(sourceEvents, evnt)
		}
	}
	tp.Rules.Events = sourceEvents

	// Convert Rules to JSON
	rulesJSON, err := json.Marshal(tp.Rules)
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
	meta := m.(SegmentMetadata)

	tpID := d.Id()
	if d.HasChanges("display_name", "rules_json_file", "import_from") {
		displayName := d.Get("display_name").(string)

		// Read rules from json file
		var tpRules segment.RuleSet
		if rulesJSON, ok := d.GetOk("rules_json_file"); ok {
			if err := json.Unmarshal([]byte(rulesJSON.(string)), &tpRules); err != nil {
				return diag.FromErr(err)
			}
		}

		// Read event library rules
		var eventLibs []segment.RuleSet
		if eventLibsIntfcs, ok := d.GetOk("import_from"); ok {
			var eventLibsJSONArray = eventLibsIntfcs.([]interface{})
			var err error
			if eventLibs, err = readEventLibs(eventLibsJSONArray); err != nil {
				return diag.FromErr(err)
			}
		}

		// Flatten event libraries
		eventLibsFlat := flattenEventLibs(eventLibs)

		// Merge json schema events with ones from the event library
		tpRules.Events = mergeEvents(eventLibsFlat.Events, tpRules.Events)

		// Construct the tracking plan
		tp := segment.TrackingPlan{
			DisplayName: displayName,
			Rules:       tpRules,
		}
		_, err := meta.client.UpdateTrackingPlan(tpID, tp)
		if err != nil {
			return diag.FromErr(err)
		}

		return resourceTrackingPlanRead(ctx, d, m)
	}
	return diags
}

func resourceTrackingPlanDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	meta := m.(SegmentMetadata)

	err := meta.client.DeleteTrackingPlan(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")

	return resourceTrackingPlanRead(ctx, d, m)
}

func flattenEventLibs(eventLibs []segment.RuleSet) segment.RuleSet {
	var eventLibsFlat segment.RuleSet
	for _, evtLib := range eventLibs {
		eventLibsFlat.Events = append(eventLibsFlat.Events, evtLib.Events...)
	}
	return eventLibsFlat
}

func readEventLibs(eventLibsJSONSlice []interface{}) ([]segment.RuleSet, error) {
	var decodedEventLibs []segment.RuleSet
	for _, eventLibJSON := range eventLibsJSONSlice {
		var eventLib segment.RuleSet
		if err := json.Unmarshal([]byte(eventLibJSON.(string)), &eventLib); err != nil {
			return nil, err
		}
		decodedEventLibs = append(decodedEventLibs, eventLib)
	}
	return decodedEventLibs, nil
}

func mergeEvents(evtLibEvents []segment.Event, tpEvents []segment.Event) []segment.Event {
	var mergedEvents []segment.Event = evtLibEvents
	for _, ruleEvnt := range tpEvents {
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
	return mergedEvents
}
