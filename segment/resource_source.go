package segment

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/uswitch/segment-config-go/segment"
)

const (
	configAllowUnplannedTrackEventsKey = "allow_unplanned_track_events"
)

var (
	allowedTrackBehaviours            ValuesSet = map[string]bool{"ALLOW": true, "OMIT_PROPERTIES": true, "BLOCK": true}
	allowedIdentifyAndGroupBehaviours ValuesSet = map[string]bool{"ALLOW": true, "OMIT_TRAITS": true, "BLOCK": true}
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
			"config": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						configAllowUnplannedTrackEventsKey: {
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
							ValidateFunc: validateCommonEventBehaviour(allowedTrackBehaviours),
						},
						"common_identify_event_on_violations": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validateCommonEventBehaviour(allowedIdentifyAndGroupBehaviours),
						},
						"common_group_event_on_violations": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validateCommonEventBehaviour(allowedIdentifyAndGroupBehaviours),
						},
					},
				},
			},
		},
		CreateContext: resourceSegmentSourceCreate,
		ReadContext:   resourceSegmentSourceRead,
		DeleteContext: resourceSegmentSourceDelete,
		UpdateContext: resourceSegmentSourceUpdate,
		// Importer: &schema.ResourceImporter{
		// 	State: resourceSegmentSourceImport,
		// },
	}
}

func resourceSegmentSourceRead(_ context.Context, r *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*segment.Client)
	id := r.Id()

	s, err := client.GetSource(id)
	if err != nil {
		return diag.FromErr(err)
	}

	r.Set("catalog_name", s.CatalogName)

	return nil
}

func resourceSegmentSourceCreate(ctx context.Context, r *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*segment.Client)
	srcName := r.Get("source_name").(string)
	catName := r.Get("catalog_name").(string)

	_, err := client.CreateSource(srcName, catName)
	if err != nil {
		log.Printf("[ERROR] Problem creating source %s \n", err)
		return diag.FromErr(err)
	}

	r.SetId(srcName)

	return resourceSegmentSourceRead(ctx, r, m)
}

func resourceSegmentSourceUpdate(ctx context.Context, r *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*segment.Client)
	srcName := r.Get("source_name").(string)
	catName := r.Get("config").(string)

	source, err := client.CreateSource(srcName, catName)
	if err != nil {
		log.Printf("[ERROR] Problem creating source %s \n", err)
		return diag.FromErr(err)
	}

	r.SetId(source.Name)

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

type ValuesSet map[string]bool

func (s ValuesSet) isAllowed(value string) bool {
	_, ok := s[value]
	return ok
}

func (s ValuesSet) values() []string {
	values := make([]string, 0, len(s))
	for k, _ := range s {
		values = append(values, k)
	}
	return values
}

type InvalidPropertyValueError struct {
	Property      string
	Value         string
	AllowedValues []string
}

func (err *InvalidPropertyValueError) Error() string {
	return fmt.Sprintf("Invalid value %s for property %s. Allowed values are: %s", err.Value, err.Property, strings.Join(err.AllowedValues, ", "))
}

func validateCommonEventBehaviour(allowedList ValuesSet) func(val interface{}, key string) (warns []string, errs []error) {
	return func(val interface{}, key string) (warns []string, errs []error) {
		if value, ok := val.(string); ok && allowedList.isAllowed(value) {
			return []string{}, []error{}
		}
		return []string{}, []error{
			&InvalidPropertyValueError{
				Property:      key,
				Value:         fmt.Sprintf("%v", val),
				AllowedValues: allowedList.values(),
			},
		}
	}
}
