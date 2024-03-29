package provider

import (
	"context"
	"encoding/json"
	"log"
	"regexp"
	"sort"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/uswitch/segment-config-go/segment"
	"github.com/uswitch/terraform-provider-segment/internal/utils"
)

func resourceTrackingPlan() *schema.Resource {
	return &schema.Resource{
		Description:   "A Segment tracking plan which outlines the events and properties to be collected across Segment Sources. More information can be found in the [Tracking Plans documentation](https://segment.com/docs/protocols/tracking-plan/create/).",
		CreateContext: resourceTrackingPlanCreate,
		ReadContext:   resourceTrackingPlanRead,
		UpdateContext: resourceTrackingPlanUpdate,
		DeleteContext: resourceTrackingPlanDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 2,
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "The ID of the tracking plan. *(Set by Segment).*",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"display_name": {
				Description: "The name of the tracking plan in the Segment UI.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"rules_json_file": {
				Description:      "The location of the JSON schema file for this tracking plan. The `file()` function should be used together with an absolute or relative path to the file.",
				Type:             schema.TypeString,
				Optional:         true,
				ValidateFunc:     validation.StringIsJSON,
				DiffSuppressFunc: utils.DiffRulesJSONState,
			},
			"import_from": {
				Description:      "The libraries to inherit events from, if any. Event libraries are defined as `data sources` and the `jsonencode([ ... ])` and `jsondecode()` functions should be used when including them, as per example above.",
				Type:             schema.TypeString,
				Optional:         true,
				ValidateDiagFunc: validateEventLibConfig,
				DiffSuppressFunc: utils.DiffRulesJSONState,
			},
			"create_time": {
				Description: "The creation time of the tracking plan. *(Set by Segment).*",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
			},
			"update_time": {
				Description: "The last time the tracking plan was updated. *(Set by Segment).*",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
			},
		},
		StateUpgraders: []schema.StateUpgrader{
			TpV1V2Upgrader(),
		},
	}
}

func validateEventLibConfig(i interface{}, _ cty.Path) diag.Diagnostics {
	if _, err := readEventLibs(i, true); err != nil {
		return diag.FromErr(err)
	}
	return nil
}

func resourceTrackingPlanCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	meta := m.(ProviderMetadata)
	client := meta.Client

	// Read tracking plan rules
	var tpRules segment.RuleSet
	if rulesJSON, ok := d.GetOk("rules_json_file"); ok {
		if err := json.Unmarshal([]byte(rulesJSON.(string)), &tpRules); err != nil {
			return diag.FromErr(err)
		}
	}

	// Read event library rules
	eventLibs, err := readEventLibs(d.GetOk("import_from"))
	if err != nil {
		return diag.FromErr(err)
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
	response, err := client.CreateTrackingPlan(tp)
	if err != nil {
		return diag.FromErr(err)
	}

	// SetId shoud utilise the calculated name part in the schema
	re := regexp.MustCompile(`rs_.*$`)
	trackingPlanID := re.FindString(response.Name)
	d.SetId(trackingPlanID)

	return resourceTrackingPlanRead(ctx, d, m)
}

func resourceTrackingPlanRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Println("[INFO] Reading tracking plans")
	var diags diag.Diagnostics
	meta := m.(ProviderMetadata)
	client := meta.Client

	tp, err := client.GetTrackingPlan(d.Id())
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
	log.Println("[INFO] Extracting event libs")
	eventLibs, err := readEventLibs(d.GetOk("import_from"))
	if err != nil {
		log.Println("[ERROR] Libs extraction failed")
		return diag.FromErr(err)
	}

	// Flatten event libraries
	eventLibsFlat := flattenEventLibs(eventLibs)

	// Remove library events
	var sourceEvents []segment.Event
	libsCount := len(eventLibsFlat.Events)
	log.Printf("[INFO] Searching %d lib events", libsCount)
	for _, evnt := range tp.Rules.Events {
		found := utils.Search(libsCount, func(i int) bool {
			log.Printf("[INFO] comparing index %d -  %s == %s", i, evnt.Name, eventLibsFlat.Events[i].Name)
			isLibEvent := evnt.Name == eventLibsFlat.Events[i].Name

			if isLibEvent {
				updateEvent(eventLibs, evnt)
			}

			return isLibEvent
		})
		log.Printf("[INFO] %s found at %d", evnt.Name, found)
		if found < 0 {
			sourceEvents = append(sourceEvents, evnt)
		}
	}

	log.Printf("[INFO] Found %d source events", len(sourceEvents))

	log.Println("[INFO] Updating event libs state from remote")
	var libsConfig string
	if eventLibs != nil {
		rawLibs, err := json.Marshal(eventLibs)
		if err != nil {
			return diag.FromErr(err)
		}
		libsConfig = string(rawLibs)
	}
	if err = d.Set("import_from", libsConfig); err != nil {
		return diag.FromErr(err)
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
	meta := m.(ProviderMetadata)
	client := meta.Client

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
		eventLibs, err := readEventLibs(d.GetOk("import_from"))
		if err != nil {
			return diag.FromErr(err)
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
		_, err = client.UpdateTrackingPlan(tpID, tp)
		if err != nil {
			return diag.FromErr(err)
		}

		return resourceTrackingPlanRead(ctx, d, m)
	}
	return diags
}

func resourceTrackingPlanDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	meta := m.(ProviderMetadata)
	client := meta.Client

	err := client.DeleteTrackingPlan(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")

	return resourceTrackingPlanRead(ctx, d, m)
}

func flattenEventLibs(eventLibs []segment.RuleSet) segment.RuleSet {
	log.Println("[INFO] Flattening event libs")
	var eventLibsFlat segment.RuleSet
	for _, evtLib := range eventLibs {
		eventLibsFlat.Events = append(eventLibsFlat.Events, evtLib.Events...)
	}

	sort.Slice(eventLibsFlat.Events, func(i, j int) bool {
		return eventLibsFlat.Events[i].Name < eventLibsFlat.Events[j].Name
	})
	return eventLibsFlat
}

func readEventLibs(eventLibsIntfcs interface{}, ok bool) ([]segment.RuleSet, error) {
	log.Println("[INFO] Reading event libs")
	if !ok || eventLibsIntfcs == nil {
		log.Println("[INFO] No libs defined")
		return nil, nil
	}

	var decodedEventLibs []segment.RuleSet
	if err := json.Unmarshal([]byte(eventLibsIntfcs.(string)), &decodedEventLibs); err != nil {
		return nil, err
	}

	log.Println("[INFO] Successfully read event libs")
	return decodedEventLibs, nil
}

func updateEvent(events []segment.RuleSet, event segment.Event) {
	for _, rs := range events {
		for i, ev := range rs.Events {
			if ev.Name == event.Name {
				rs.Events[i] = event
			}
		}
	}
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

// State migrations

// V1 -> V2

func readEventLibsV1(eventLibsIntfcs interface{}, ok bool) ([]segment.RuleSet, error) {
	log.Println("[INFO] Reading v1 event libs")
	if !ok || eventLibsIntfcs == nil {
		log.Println("[INFO] No libs defined")
		return nil, nil
	}

	eventLibsJSONSlice := eventLibsIntfcs.([]interface{})
	decodedEventLibs := []segment.RuleSet{}
	for _, eventLibJSON := range eventLibsJSONSlice {
		var eventLib segment.RuleSet
		if err := json.Unmarshal([]byte(eventLibJSON.(string)), &eventLib); err != nil {
			return nil, err
		}
		decodedEventLibs = append(decodedEventLibs, eventLib)
	}
	return decodedEventLibs, nil
}

func tpResourceV1() *schema.Resource {
	return &schema.Resource{
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
				DiffSuppressFunc: utils.DiffRulesJSONState,
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

func TpV1V2Upgrader() schema.StateUpgrader {
	return schema.StateUpgrader{
		Type: tpResourceV1().CoreConfigSchema().ImpliedType(),
		Upgrade: func(_ context.Context, rawState map[string]interface{}, _ interface{}) (map[string]interface{}, error) {
			log.Println("[INFO] Migrating Schema V1 -> V2")
			value, ok := rawState["import_from"]

			old, err := readEventLibsV1(value, ok)
			if err != nil {
				log.Println("[INFO] Migration V1 -> V2 failed")
				return nil, err
			}

			if old == nil {
				log.Println("[INFO] import_from not set, skipping")
				return rawState, nil
			}

			libsJSON, err := json.Marshal(old)
			if err != nil {
				log.Println("[INFO] Migration V1 -> V2 failed")
				return nil, err
			}

			rawState["import_from"] = string(libsJSON)

			log.Println("[INFO] Successfully migrated V1 -> V2")
			return rawState, nil
		},
		Version: 1,
	}
}
