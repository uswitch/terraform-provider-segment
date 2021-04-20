package segment

import (
	"context"
	"fmt"

	"github.com/ajbosco/segment-config-go/segment"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
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
		AllowTrackEventOnViolations:         true,
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
			"source_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"catalog_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"schema_config": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
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

	config, err := client.GetSourceConfig(id)
	if err != nil {
		return diag.FromErr(err)
	}

	err = r.Set("catalog_name", s.CatalogName)
	if err != nil {
		return diag.FromErr(err)
	}

	err = r.Set("source_name", id)
	if err != nil {
		return diag.FromErr(err)
	}

	if _, exists := r.GetOk("schema_config"); exists {
		err = r.Set("schema_config", encodeSourceConfig(config))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	return nil
}

func resourceSegmentSourceCreate(ctx context.Context, r *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*segment.Client)
	srcName := r.Get("source_name").(string)
	catName := r.Get("catalog_name").(string)
	configs := r.Get("schema_config").([]interface{})
	var config *segment.SourceConfig

	if len(configs) > 0 {
		var err *diag.Diagnostics

		var c segment.SourceConfig
		if err = decodeSourceConfig(configs[0], &c); err != nil {
			return *err
		}
		config = &c
	}

	if _, err := client.CreateSource(srcName, catName); err != nil {
		return diag.FromErr(err)
	}

	// ignoring when config is not set
	if config != nil {
		if s, err := client.UpdateSourceConfig(srcName, *config); err != nil {
			// Reverting resource creation on failure
			diags := diag.FromErr(err)
			if err := client.DeleteSource(srcName); err != nil {
				diags = append(diags, diag.Diagnostic{
					Severity: diag.Warning,
					Summary:  "Lingering Segment resources",
					Detail:   fmt.Sprintf("Source %s could not be cleaned up because of %s. Check Segment for manual cleanup", s.Name, err),
				})
			}
			return diags
		}
	}

	r.SetId(srcName)

	return resourceSegmentSourceRead(ctx, r, m)
}

func resourceSegmentSourceUpdate(ctx context.Context, r *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*segment.Client)
	srcName := r.Get("source_name").(string)

	if r.HasChange("schema_config") {

		configs := r.Get("schema_config").([]interface{})
		var config segment.SourceConfig

		if len(configs) > 0 {
			if d := decodeSourceConfig(configs[0], &config); d != nil {
				return *d
			}
		} else {
			config = defaultSourceConfig
		}

		_, err := client.UpdateSourceConfig(srcName, config)
		if err != nil {
			return diag.FromErr(err)
		}
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
			err := diag.FromErr(fmt.Errorf("failed to decode schema config into a valid SourceConfig: %w", r.(error)))
			diags = &err
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
