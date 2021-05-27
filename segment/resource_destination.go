package segment

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/ajbosco/segment-config-go/segment"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceSegmentDestination() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"source": {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool { return true },
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"enabled": {
				Type:     schema.TypeBool,
				Required: true,
			},
			"parent": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"display_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"create_time": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"update_time": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"connection_mode": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringInSlice([]string{"UNSPECIFIED", "CLOUD", "DEVICE"}, false),
			},
			"config": {
				Type:     schema.TypeMap,
				Required: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
		CreateContext: resourceSegmentDestinationCreate,
		ReadContext:   resourceSegmentDestinationRead,
		UpdateContext: resourceSegmentDestinationUpdate,
		DeleteContext: resourceSegmentDestinationDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

func idToSourceAndDest(r *schema.ResourceData) (string, string) {
	id := r.Id()
	parts := strings.Split(id, "/")

	if len(parts) > 2 {
		panic("Invalid destination id: " + id)
	}

	return parts[0], parts[1]
}

func destinationResourceId(src string, dst string) string {
	return src + "/" + "dst"
}

func resourceSegmentDestinationRead(_ context.Context, r *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*segment.Client)
	srcName, dstName := idToSourceAndDest(r)

	d, err := client.GetDestination(srcName, dstName)
	if err != nil {
		return diag.FromErr(err)
	}

	config := map[string]interface{}{}
	if err := encodeDestinationConfig(d, &config); err != nil {
		return *err
	}

	r.Set("source", srcName)
	r.Set("name", pathToName(d.Name))
	r.Set("enabled", d.Enabled)
	r.Set("parent", d.Parent)
	r.Set("display_name", d.DisplayName)
	r.Set("connection_mode", d.ConnectionMode)
	r.Set("config", config)
	r.Set("create_time", d.CreateTime.Unix())
	r.Set("update_time", d.UpdateTime.Unix())

	return nil
}

func resourceSegmentDestinationUpdate(ctx context.Context, r *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*segment.Client)
	srcName := r.Get("source").(string)
	destName := r.Get("name").(string)
	enabled := r.Get("enabled").(bool)

	config := []segment.DestinationConfig{}
	if d := decodeDestinationConfig("uswitch-sandbox", srcName, destName, r.Get("config"), &config); d != nil {
		return *d
	}

	if _, err := client.UpdateDestination(srcName, destName, enabled, config); err != nil {
		return diag.FromErr(err)
	}

	return resourceSegmentDestinationRead(ctx, r, m)
}

func resourceSegmentDestinationCreate(ctx context.Context, r *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*segment.Client)
	srcName := r.Get("source").(string)
	destName := r.Get("name").(string)
	id := destinationResourceId(srcName, destName)

	mode := r.Get("connection_mode").(string)
	enabled := r.Get("enabled").(bool)
	var config []segment.DestinationConfig
	if d := decodeDestinationConfig("uswitch-sandbox", srcName, destName, r.Get("config"), &config); d != nil {
		return *d
	}

	if _, err := client.CreateDestination(srcName, destName, mode, enabled, config); err != nil {
		return diag.FromErr(err)
	}

	r.SetId(id)

	return resourceSegmentDestinationRead(ctx, r, m)
}

func resourceSegmentDestinationDelete(_ context.Context, r *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*segment.Client)
	srcName := r.Get("source").(string)
	destName := r.Get("name").(string)

	err := client.DeleteDestination(srcName, destName)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

// Decoders

func encodeDestinationConfig(destination segment.Destination, encoded *map[string]interface{}) *diag.Diagnostics {
	if encoded == nil {
		return diagFromErrPtr(errors.New("destination config encoded map cannot be nil"))
	}

	for _, config := range destination.Configs {
		c, err := json.Marshal(map[string]interface{}{
			"type":  config.Type,
			"value": config.Value,
		})
		if err != nil {
			diagFromErrPtr(err)
		}

		(*encoded)[pathToName(config.Name)] = string(c)
	}

	return nil
}

func decodeDestinationConfig(workspace string, srcName string, destName string, rawConfig interface{}, dst *[]segment.DestinationConfig) (diags *diag.Diagnostics) {
	defer func() {
		if r := recover(); r != nil {
			diags = diagFromErrPtr(fmt.Errorf("failed to decode destination config: %w", r.(error)))
		}
	}()

	configs := rawConfig.(map[string]interface{})
	for k, configRaw := range configs {
		var config segment.DestinationConfig
		if json.Unmarshal([]byte(configRaw.(string)), &config) != nil {
			panic(fmt.Errorf("invalid config value: %s", configRaw))
		}

		// Version will error on update to the Config API
		if k == "version" {
			continue
		}

		config.Name = fmt.Sprintf("workspaces/%s/sources/%s/destinations/%s/config/%s", workspace, srcName, destName, k)

		log.Printf("CONFIG: %v", config)

		*dst = append(*dst, config)
	}

	return nil
}

// Misc Helpers
