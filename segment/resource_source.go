package segment

import (
	"context"
	"fmt"
	"log"
	"reflect"
	"strings"
	"time"

	"github.com/ajbosco/segment-config-go/segment"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

const (
	keySource       = "source_name"
	keyCatalog      = "catalog_name"
	keyTrackingPlan = "tracking_plan"
	keySchemaConfig = "schema_config"
)

var (
	allowedTrackBehaviours            = []string{"ALLOW", "OMIT_PROPERTIES", "BLOCK"}
	allowedIdentifyAndGroupBehaviours = []string{"ALLOW", "OMIT_TRAITS", "BLOCK"}
	defaultSourceConfig               = segment.SourceConfig{
		AllowUnplannedTrackEvents:           true,
		AllowUnplannedIdentifyTraits:        true,
		AllowUnplannedGroupTraits:           true,
		ForwardingBlockedEventsTo:           "",
		AllowUnplannedTrackEventsProperties: true,
		AllowTrackEventOnViolations:         false,
		AllowIdentifyTraitsOnViolations:     true,
		AllowGroupTraitsOnViolations:        true,
		ForwardingViolationsTo:              "",
		AllowTrackPropertiesOnViolations:    true,
		CommonTrackEventOnViolations:        segment.Allow,
		CommonIdentifyEventOnViolations:     segment.Allow,
		CommonGroupEventOnViolations:        segment.Allow,
	}
)

func resourceSegmentSource() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			keySource: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			keyCatalog: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			keyTrackingPlan: {
				Type:     schema.TypeString,
				Optional: true,
			},
			keySchemaConfig: {
				Type:             schema.TypeList,
				Optional:         true,
				MaxItems:         1,
				RequiredWith:     []string{keyTrackingPlan},
				DiffSuppressFunc: suppressSchemaConfigDiff,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"allow_unplanned_track_events": {
							Type:     schema.TypeBool,
							Required: true,
						},
						"allow_unplanned_identify_traits": {
							Type:     schema.TypeBool,
							Required: true,
						},
						"allow_unplanned_group_traits": {
							Type:     schema.TypeBool,
							Required: true,
						},
						"forwarding_blocked_events_to": {
							Type:     schema.TypeString,
							Optional: true,
							Default:  "",
						},
						"allow_unplanned_track_event_properties": {
							Type:     schema.TypeBool,
							Required: true,
						},
						"allow_track_event_on_violations": {
							Type:     schema.TypeBool,
							Required: true,
						},
						"allow_identify_traits_on_violations": {
							Type:     schema.TypeBool,
							Required: true,
						},
						"allow_group_traits_on_violations": {
							Type:     schema.TypeBool,
							Required: true,
						},
						"forwarding_violations_to": {
							Type:     schema.TypeString,
							Optional: true,
							Default:  "",
						},
						"allow_track_properties_on_violations": {
							Type:     schema.TypeBool,
							Required: true,
						},
						"common_track_event_on_violations": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice(allowedTrackBehaviours, false),
						},
						"common_identify_event_on_violations": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice(allowedIdentifyAndGroupBehaviours, false),
						},
						"common_group_event_on_violations": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice(allowedIdentifyAndGroupBehaviours, false),
						},
						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"parent": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
		CreateContext: resourceSegmentSourceCreate,
		ReadContext:   resourceSegmentSourceRead,
		DeleteContext: resourceSegmentSourceDelete,
		UpdateContext: resourceSegmentSourceUpdate,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

func resourceSegmentSourceRead(_ context.Context, r *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*segment.Client)
	id := r.Id()

	s, err := client.GetSource(id)
	if err != nil {
		return diag.FromErr(err)
	}

	if err = r.Set(keyCatalog, s.CatalogName); err != nil {
		return diag.FromErr(err)
	}

	if err = r.Set(keySource, id); err != nil {
		return diag.FromErr(err)
	}

	tpID, d := initTrackingPlan(r.Get(keyTrackingPlan).(string), id, *client)
	if d != nil {
		return *d
	}

	hasTrackingPlanSet := tpID != ""
	if hasTrackingPlanSet {
		if err = r.Set(keyTrackingPlan, tpID); err != nil {
			return diag.FromErr(err)
		}

		config, err := client.GetSourceConfig(id)
		if err != nil {
			return diag.FromErr(err)
		}
		if err = r.Set(keySchemaConfig, encodeSourceConfig(config)); err != nil {
			return diag.FromErr(err)
		}
	}

	return nil
}

func resourceSegmentSourceCreate(ctx context.Context, r *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*segment.Client)
	srcName := r.Get(keySource).(string)
	catName := r.Get(keyCatalog).(string)

	if _, err := client.CreateSource(srcName, catName); err != nil {
		return diag.FromErr(err)
	}

	revertCreation := func(d diag.Diagnostics) diag.Diagnostics {
		if err := client.DeleteSource(srcName); err != nil {
			d = append(d, diag.Diagnostic{
				Severity: diag.Warning,
				Summary:  "Lingering Segment resources",
				Detail:   fmt.Sprintf("Source %s could not be cleaned up because of %s. Check Segment for manual cleanup", srcName, err),
			})
		}
		return d
	}

	if d := updateTrackingPlan(r, *client); d != nil {
		return revertCreation(*d)
	}

	if d := updateSchemaConfig(r, *client); d != nil {
		return revertCreation(*d)
	}

	r.SetId(srcName)

	return resourceSegmentSourceRead(ctx, r, m)
}

func resourceSegmentSourceUpdate(ctx context.Context, r *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*segment.Client)
	srcName := r.Get(keySource).(string)

	if d := updateTrackingPlan(r, *client); d != nil {
		return *d
	}

	if d := updateSchemaConfig(r, *client); d != nil {
		return *d
	}

	r.SetId(srcName)

	return resourceSegmentSourceRead(ctx, r, m)
}

func resourceSegmentSourceDelete(_ context.Context, r *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*segment.Client)
	id := r.Id()

	err := client.DeleteSource(id)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

// Decoders

func encodeSourceConfig(config segment.SourceConfig) []map[string]interface{} {
	return []map[string]interface{}{{
		"allow_unplanned_track_events":           config.AllowUnplannedTrackEvents,
		"allow_unplanned_identify_traits":        config.AllowUnplannedIdentifyTraits,
		"allow_unplanned_group_traits":           config.AllowUnplannedGroupTraits,
		"forwarding_blocked_events_to":           config.ForwardingBlockedEventsTo,
		"allow_unplanned_track_event_properties": config.AllowUnplannedTrackEventsProperties,
		"allow_track_event_on_violations":        config.AllowTrackEventOnViolations,
		"allow_identify_traits_on_violations":    config.AllowIdentifyTraitsOnViolations,
		"allow_group_traits_on_violations":       config.AllowGroupTraitsOnViolations,
		"forwarding_violations_to":               config.ForwardingViolationsTo,
		"allow_track_properties_on_violations":   config.AllowTrackPropertiesOnViolations,
		"common_track_event_on_violations":       config.CommonTrackEventOnViolations,
		"common_identify_event_on_violations":    config.CommonIdentifyEventOnViolations,
		"common_group_event_on_violations":       config.CommonGroupEventOnViolations,
		"name":                                   config.Name,
		"parent":                                 config.Parent,
	}}
}

func decodeSourceConfig(rawConfigMap interface{}, dst *segment.SourceConfig) (diags *diag.Diagnostics) {
	defer func() {
		if r := recover(); r != nil {
			diags = diagFromErrPtr(fmt.Errorf("failed to decode schema config into a valid SourceConfig: %w", r.(error)))
		}
	}()

	configMap := rawConfigMap.(map[string]interface{})
	*dst = segment.SourceConfig{
		AllowUnplannedTrackEvents:           configMap["allow_unplanned_track_events"].(bool),
		AllowUnplannedIdentifyTraits:        configMap["allow_unplanned_identify_traits"].(bool),
		AllowUnplannedGroupTraits:           configMap["allow_unplanned_group_traits"].(bool),
		ForwardingBlockedEventsTo:           configMap["forwarding_blocked_events_to"].(string),
		AllowUnplannedTrackEventsProperties: configMap["allow_unplanned_track_event_properties"].(bool),
		AllowTrackEventOnViolations:         configMap["allow_track_event_on_violations"].(bool),
		AllowIdentifyTraitsOnViolations:     configMap["allow_identify_traits_on_violations"].(bool),
		AllowGroupTraitsOnViolations:        configMap["allow_group_traits_on_violations"].(bool),
		ForwardingViolationsTo:              configMap["forwarding_violations_to"].(string),
		AllowTrackPropertiesOnViolations:    configMap["allow_track_properties_on_violations"].(bool),
		CommonTrackEventOnViolations:        segment.CommonEventSettings(configMap["common_track_event_on_violations"].(string)),
		CommonIdentifyEventOnViolations:     segment.CommonEventSettings(configMap["common_identify_event_on_violations"].(string)),
		CommonGroupEventOnViolations:        segment.CommonEventSettings(configMap["common_group_event_on_violations"].(string)),
	}

	return
}

// Schema config

func getSchemaConfigOrDefault(r *schema.ResourceData, dst *segment.SourceConfig) *diag.Diagnostics {
	configs := r.Get(keySchemaConfig).([]interface{})
	hasConfigSet := len(configs) > 0

	if hasConfigSet {
		if d := decodeSourceConfig(configs[0], dst); d != nil {
			return d
		}
	} else {
		*dst = defaultSourceConfig
	}

	return nil
}

func updateSchemaConfig(r *schema.ResourceData, client segment.Client) *diag.Diagnostics {
	if !r.HasChange(keySchemaConfig) {
		return nil
	}

	srcName := r.Get(keySource).(string)
	tpID := r.Get(keyTrackingPlan).(string)
	hasTrackingPlanSet := tpID != ""

	if hasTrackingPlanSet {
		var config segment.SourceConfig
		if d := getSchemaConfigOrDefault(r, &config); d != nil {
			return d
		}

		_, err := client.UpdateSourceConfig(srcName, config)
		if err != nil {
			return diagFromErrPtr(err)
		}
	} else {
		// We wipe the previous config out of the state as there's no more tracking plan attached
		r.Set(keySchemaConfig, nil)
	}

	return nil
}

// Tracking Plans

func updateTrackingPlan(r *schema.ResourceData, client segment.Client) *diag.Diagnostics {
	srcName := r.Get(keySource).(string)

	if old, new := r.GetChange(keyTrackingPlan); old != new {
		if old != "" {
			if err := client.DeleteTrackingPlanSourceConnection(old.(string), srcName); err != nil {
				return diagFromErrPtr(err)
			}
		}

		if new != "" {
			if err := client.CreateTrackingPlanSourceConnection(new.(string), srcName); err != nil {
				return diagFromErrPtr(err)
			}
		}
	}
	return nil
}

func initTrackingPlan(tpID string, source string, client segment.Client) (string, *diag.Diagnostics) {
	if tpID != "" {
		// We first try to match the tracking plan specified in the config to avoid expensive calls
		if d := assertTrackingPlanConnected(tpID, source, client); d != nil {
			return findTrackingPlanSourceConnection(source, client)
		}

		return tpID, nil
	} else {
		// When the tracking plan is not specified, we search for it, so we can import existing sources
		return findTrackingPlanSourceConnection(source, client)
	}
}

// assertTrackingPlanConnected verifies a tracking plan and a source are connected and fails otherwise
func assertTrackingPlanConnected(trackingPlan string, src string, client segment.Client) *diag.Diagnostics {
	sources, err := client.ListTrackingPlanSources(trackingPlan)
	if err != nil {
		return diagFromErrPtr(fmt.Errorf("invalid tracking plan ID %s: %w", trackingPlan, err))
	}

	for _, s := range sources {
		if pathToName(s.Source) == src {
			return nil
		}
	}

	d := diag.Errorf("Tracking plan not found: %s", trackingPlan)
	return &d
}

// findTrackingPlanSourceConnection finds the connected tracking plan, or "" if the source is not connected
func findTrackingPlanSourceConnection(source string, client segment.Client) (string, *diag.Diagnostics) {
	// We have to browse all tracking plans to find the source.
	// Ideally we'll have the tracking plan attached to the source or fetchable through a unique endpoint in the future
	tps, err := client.ListTrackingPlans()
	if err != nil {
		return "", diagFromErrPtr(err)
	}

	for _, tp := range tps.TrackingPlans {
		tpID := pathToName(tp.Name)
		srcs, err := client.ListTrackingPlanSources(tpID)
		if err != nil {
			return "", diagFromErrPtr(err)
		}

		for _, currSrc := range srcs {
			if pathToName(currSrc.Source) == source {
				return tpID, nil
			}
		}
		time.Sleep(75 * time.Millisecond)
	}

	return "", nil
}

// Misc Helpers

func diagFromErrPtr(err error) *diag.Diagnostics {
	d := diag.FromErr(err)
	return &d
}

// Converts a segment resource path to its id
// E.g: workspaces/myworkspace/sources/mysource => mysource
func pathToName(path string) string {
	parts := strings.Split(path, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}

	return path
}

// suppressSchemaConfigDiff hides changes to schema config when it is not specified explicitely but using the default one
func suppressSchemaConfigDiff(k, old, new string, d *schema.ResourceData) bool {
	var config segment.SourceConfig
	if err := decodeSourceConfig(d.Get(keySchemaConfig).([]interface{})[0], &config); err != nil {
		log.Printf("[WARNING]: Problem when suppressing diff for %s: %s => %s: %v", k, old, new, err)
		return false
	}

	noChange := !d.HasChange(keySchemaConfig)
	isUsingDefaultConfig := reflect.DeepEqual(config, defaultSourceConfig)
	return noChange && isUsingDefaultConfig
}
