package segment

import (
	"context"
	"fmt"
	"math"
	"path"
	"strings"

	"github.com/ajbosco/segment-config-go/segment"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/uswitch/terraform-provider-segment/segment/internal/utils"
)

const (
	keyFilterDestination      = "destination"
	keyFilterName             = "name"
	keyFilterTitle            = "title"
	keyFilterDescription      = "description"
	keyFilterCondition        = "condition"
	keyFilterEnabled          = "enabled"
	keyFilterActions          = "actions"
	keyFilterActionDrop       = "drop"
	keyFilterActionBlock      = "block_fields"
	keyFilterActionAllow      = "allow_fields"
	keyFilterActionSample     = "sample"
	keyFilterActionContext    = "context"
	keyFilterActionTraits     = "traits"
	keyFilterActionProperties = "properties"
	keyFilterActionPercent    = "percent"
	keyFilterActionPath       = "path"
)

var eventFilterActionSchema = schema.Schema{
	Type:          schema.TypeList,
	Optional:      true,
	MaxItems:      1,
	Default:       nil,
	ConflictsWith: []string{keyFilterActions + ".0." + keyFilterActionDrop + ".0"},
	Elem: &schema.Resource{
		Schema: map[string]*schema.Schema{
			keyFilterActionTraits: {
				Type: schema.TypeSet,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Optional: true,
			},
			keyFilterActionContext: {
				Type: schema.TypeSet,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Optional: true,
			},
			keyFilterActionProperties: {
				Type: schema.TypeSet,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Optional: true,
			},
		},
	},
}

func resourceSegmentDestinationFilter() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			keyFilterDestination: {
				Type:     schema.TypeString,
				Required: true,
			},
			keyFilterName: {
				Type:     schema.TypeString,
				Computed: true,
			},
			keyFilterTitle: {
				Type:     schema.TypeString,
				Required: true,
			},
			keyFilterDescription: {
				Type:     schema.TypeString,
				Required: true,
			},
			keyFilterCondition: {
				Type:     schema.TypeString,
				Required: true,
			},
			keyFilterEnabled: {
				Type:     schema.TypeBool,
				Required: true,
			},
			keyFilterActions: {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				MinItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						keyFilterActionDrop: {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{},
							},
						},
						keyFilterActionBlock: &eventFilterActionSchema,
						keyFilterActionAllow: &eventFilterActionSchema,
						keyFilterActionSample: {
							Type:          schema.TypeSet,
							Optional:      true,
							Default:       nil,
							ConflictsWith: []string{keyFilterActions + ".0." + keyFilterActionDrop + ".0"},
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									keyFilterActionPercent: {
										Type:         schema.TypeFloat,
										ValidateFunc: validation.FloatBetween(0, 1),
										Required:     true,
									},
									keyFilterActionPath: {
										Type:     schema.TypeString,
										Optional: true,
									},
								},
							},
						},
					},
				},
			},
		},
		CreateContext: resourceSegmentDestinationFilterCreate,
		ReadContext:   resourceSegmentDestinationFilterRead,
		UpdateContext: resourceSegmentDestinationFilterUpdate,
		DeleteContext: resourceSegmentDestinationFilterDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

func resourceSegmentDestinationFilterRead(_ context.Context, r *schema.ResourceData, m interface{}) diag.Diagnostics {
	meta := m.(ProviderMetadata)
	client := meta.Client

	f, err := client.GetDestinationFilter(SplitDestinationFilterId(r.Id()))
	if err != nil {
		return diag.FromErr(err)
	}

	return utils.CatchFirst(
		func() error { return r.Set(keyFilterTitle, f.Title) },
		func() error { return r.Set(keyFilterDescription, f.Description) },
		func() error { return r.Set(keyFilterEnabled, f.IsEnabled) },
		func() error { return r.Set(keyFilterCondition, f.Conditions) },
		func() error { return r.Set(keyFilterName, f.Name) },
		func() error { return r.Set(keyFilterActions, encodeDestinationFilterActions(f.Actions)) },
	)
}

func resourceSegmentDestinationFilterUpdate(ctx context.Context, r *schema.ResourceData, m interface{}) diag.Diagnostics {
	meta := m.(ProviderMetadata)
	client := meta.Client
	src, dest, _ := SplitDestinationFilterId(r.Id())

	var filter segment.DestinationFilter
	if d := decodeDestinationFilter(r, &filter); d != nil {
		return d
	}

	if _, err := client.UpdateDestinationFilter(src, dest, filter); err != nil {
		return diag.FromErr(err)
	}

	return resourceSegmentDestinationFilterRead(ctx, r, m)
}

func resourceSegmentDestinationFilterCreate(ctx context.Context, r *schema.ResourceData, m interface{}) diag.Diagnostics {
	meta := m.(ProviderMetadata)
	client := meta.Client
	destinationId := r.Get(keyFilterDestination).(string)
	srcName, dstName := destinationIdToSourceAndDest(destinationId)

	var f segment.DestinationFilter
	if d := decodeDestinationFilter(r, &f); d != nil {
		return d
	}

	created, err := client.CreateDestinationFilter(srcName, dstName, f)
	if err != nil {
		return diag.FromErr(err)
	}

	_, id := path.Split(created.Name)
	r.SetId(destinationFilterResourceId(srcName, dstName, id))

	return resourceSegmentDestinationFilterRead(ctx, r, m)
}

func resourceSegmentDestinationFilterDelete(_ context.Context, r *schema.ResourceData, m interface{}) diag.Diagnostics {
	meta := m.(ProviderMetadata)
	client := meta.Client
	srcName, dstName, id := SplitDestinationFilterId(r.Id())

	err := client.DeleteDestinationFilter(srcName, dstName, id)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

// Decoders

func decodeDestinationFilter(r *schema.ResourceData, dst *segment.DestinationFilter) (diags diag.Diagnostics) {
	defer func() {
		if r := recover(); r != nil {
			diags = diag.FromErr(fmt.Errorf("failed to decode destination filter config: %w", r.(error)))
		}
	}()

	f := segment.DestinationFilter{
		Name:        r.Get(keyFilterName).(string),
		Title:       r.Get(keyFilterTitle).(string),
		Description: r.Get(keyFilterDescription).(string),
		Conditions:  r.Get(keyFilterCondition).(string),
		IsEnabled:   r.Get(keyFilterEnabled).(bool),
	}

	rawActions := r.Get(keyFilterActions).([]interface{})[0].(map[string]interface{})

	if len(rawActions[keyFilterActionDrop].([]interface{})) == 1 {
		f.Actions = append(f.Actions, segment.NewDropEventAction())
	}

	if rawBlocks := rawActions[keyFilterActionBlock].([]interface{}); len(rawBlocks) > 0 {
		fields := rawBlocks[0].(map[string]interface{})
		context := setAsStrSlice(fields[keyFilterActionContext])
		props := setAsStrSlice(fields[keyFilterActionProperties])
		traits := setAsStrSlice(fields[keyFilterActionTraits])
		f.Actions = append(f.Actions, segment.NewBlockListEventAction(props, context, traits))
	}

	if rawAllows := rawActions[keyFilterActionAllow].([]interface{}); len(rawAllows) > 0 {
		fields := rawAllows[0].(map[string]interface{})
		context := setAsStrSlice(fields[keyFilterActionContext])
		props := setAsStrSlice(fields[keyFilterActionProperties])
		traits := setAsStrSlice(fields[keyFilterActionTraits])
		f.Actions = append(f.Actions, segment.NewAllowListEventAction(props, context, traits))
	}

	for _, rawSample := range rawActions[keyFilterActionSample].(*schema.Set).List() {
		sample := rawSample.(map[string]interface{})
		percent := float32(sample[keyFilterActionPercent].(float64))
		path := sample[keyFilterActionPath].(string)
		f.Actions = append(f.Actions, segment.NewSamplingEventAction(percent, path))
	}

	*dst = f

	return nil
}

func encodeDestinationFilterActions(actions segment.DestinationFilterActions) []interface{} {
	if len(actions) == 0 {
		return nil
	}

	root := map[string][]interface{}{}
	for _, action := range actions {
		switch action.ActionType() {
		case segment.DestinationFilterActionTypeDropEvent:
			root[keyFilterActionDrop] = []interface{}{nil}
		case segment.DestinationFilterActionTypeBlockList:
			root[keyFilterActionBlock] = decodeEventDescription(action.(segment.FieldsListEventAction).Fields)
		case segment.DestinationFilterActionTypeAllowList:
			root[keyFilterActionAllow] = decodeEventDescription(action.(segment.FieldsListEventAction).Fields)
		case segment.DestinationFilterActionTypeSampling:
			root[keyFilterActionSample] = append(root[keyFilterActionSample], map[string]interface{}{
				// We need to round so that we don't lose precision and get a diff like
				// - 1.2345
				// + 1.23
				// when the expected was always 1.23
				keyFilterActionPercent: math.Round(float64(action.(segment.SamplingEventAction).Percent*100)) / 100,
				keyFilterActionPath:    action.(segment.SamplingEventAction).Path,
			})
		}
	}

	return []interface{}{root}
}

func decodeEventDescription(list segment.EventDescription) []interface{} {
	result := map[string]interface{}{}

	if list.Properties != nil && len(list.Properties.Fields) > 0 {
		result[keyFilterActionProperties] = strSliceToSet(list.Properties.Fields)
	}
	if list.Context != nil && len(list.Context.Fields) > 0 {
		result[keyFilterActionContext] = strSliceToSet(list.Context.Fields)
	}
	if list.Traits != nil && len(list.Traits.Fields) > 0 {
		result[keyFilterActionTraits] = strSliceToSet(list.Traits.Fields)
	}

	return []interface{}{result}
}

// Misc Helpers

func SplitDestinationFilterId(id string) (sourceName string, destinationName string, filterId string) {
	parts := strings.Split(id, "/")
	return parts[0], parts[1], parts[2]
}

func destinationFilterResourceId(s string, d string, filterId string) string {
	return s + "/" + d + "/" + filterId
}

func strSliceToSet(strs []string) *schema.Set {
	result := []interface{}{}
	for _, v := range strs {
		result = append(result, v)
	}

	return schema.NewSet(schema.HashString, result)
}

func setAsStrSlice(s interface{}) []string {
	list := []string{}
	for _, v := range s.(*schema.Set).List() {
		list = append(list, v.(string))
	}

	return list
}
