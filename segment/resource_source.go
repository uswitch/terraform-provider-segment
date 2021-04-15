package segment

import (
	"context"
	"log"

	"github.com/fatih/structs"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/uswitch/segment-config-go/segment"
)

var (
	allowedTrackBehaviours            = []string{"ALLOW", "OMIT_PROPERTIES", "BLOCK"}
	allowedIdentifyAndGroupBehaviours = []string{"ALLOW", "OMIT_TRAITS", "BLOCK"}
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

	config, err := client.GetSourceConfig(id)
	log.Printf("[DEBUG] Config: %+v \n", config)
	if err != nil {
		return diag.FromErr(err)
	}

	err = r.Set("catalog_name", s.CatalogName)
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[DEBUG] Setting config to: %+v \n")
	structs.DefaultTagName = "json"
	cfg := structs.Map(config)
	err = r.Set("config", []map[string]interface{}{cfg})
	if err != nil {
		return diag.FromErr(err)
	}

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
