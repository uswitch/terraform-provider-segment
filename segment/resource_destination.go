package segment

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/ajbosco/segment-config-go/segment"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

const (
	keyDestSource      = "source"
	keyDestName        = "name"
	keyDestEnabled     = "enabled"
	keyDestDisplayName = "display_name"
	keyDestCreateTime  = "create_time"
	keyDestUpdateTime  = "update_time"
	keyDestConMode     = "connection_mode"
	keyDestConfig      = "config"
	keyDestParent      = "parent"
)

func resourceSegmentDestination() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			keyDestSource: {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool { return true },
			},
			keyDestName: {
				Type:     schema.TypeString,
				Required: true,
			},
			keyDestEnabled: {
				Type:     schema.TypeBool,
				Required: true,
			},
			keyDestParent: {
				Type:     schema.TypeString,
				Computed: true,
			},
			keyDestDisplayName: {
				Type:     schema.TypeString,
				Computed: true,
			},
			keyDestCreateTime: {
				Type:     schema.TypeString,
				Computed: true,
			},
			keyDestUpdateTime: {
				Type:     schema.TypeInt,
				Computed: true,
			},
			keyDestConMode: {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringInSlice([]string{"UNSPECIFIED", "CLOUD", "DEVICE"}, false),
			},
			keyDestConfig: {
				Type:     schema.TypeMap,
				Required: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				ValidateDiagFunc: validateDestinationConfig,
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

func resourceSegmentDestinationRead(_ context.Context, r *schema.ResourceData, m interface{}) diag.Diagnostics {
	meta := m.(ProviderMetadata)
	client := meta.client
	srcName, dstName := idToSourceAndDest(r)

	d, err := client.GetDestination(srcName, dstName)
	if err != nil {
		return diag.FromErr(err)
	}

	config := map[string]interface{}{}
	if err := encodeDestinationConfig(d, &config); err != nil {
		return *err
	}

	if err := r.Set(keyDestSource, srcName); err != nil {
		return diag.FromErr(err)
	}
	if err := r.Set(keyDestName, pathToName(d.Name)); err != nil {
		return diag.FromErr(err)
	}
	if err := r.Set(keyDestEnabled, d.Enabled); err != nil {
		return diag.FromErr(err)
	}
	if err := r.Set(keyDestParent, d.Parent); err != nil {
		return diag.FromErr(err)
	}
	if err := r.Set(keyDestDisplayName, d.DisplayName); err != nil {
		return diag.FromErr(err)
	}
	if err := r.Set(keyDestConMode, d.ConnectionMode); err != nil {
		return diag.FromErr(err)
	}
	if err := r.Set(keyDestConfig, config); err != nil {
		return diag.FromErr(err)
	}
	if err := r.Set(keyDestCreateTime, d.CreateTime.String()); err != nil {
		return diag.FromErr(err)
	}
	if err := r.Set(keyDestUpdateTime, d.UpdateTime.String()); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceSegmentDestinationUpdate(ctx context.Context, r *schema.ResourceData, m interface{}) diag.Diagnostics {
	meta := m.(ProviderMetadata)
	client := meta.client
	srcName := r.Get(keyDestSource).(string)
	destName := r.Get(keyDestName).(string)
	enabled := r.Get(keyDestEnabled).(bool)

	log.Println("[INFO] Fetching current config for " + srcName)
	d, err := client.GetDestination(srcName, destName)
	if err != nil {
		return diag.FromErr(err)
	}

	config := []segment.DestinationConfig{}
	if d := decodeDestinationConfig(meta.workspace, srcName, destName, d.Configs, &config); d != nil {
		return *d
	}

	if _, err := client.UpdateDestination(srcName, destName, enabled, config); err != nil {
		return diag.FromErr(err)
	}

	return resourceSegmentDestinationRead(ctx, r, m)
}

func resourceSegmentDestinationCreate(ctx context.Context, r *schema.ResourceData, m interface{}) diag.Diagnostics {
	meta := m.(ProviderMetadata)
	client := meta.client
	srcName := r.Get(keyDestSource).(string)
	destName := r.Get(keyDestName).(string)
	id := destinationResourceId(srcName, destName)

	log.Println("[INFO] Fetching current config for " + srcName)
	_, err := client.GetDestination(srcName, destName)
	if err != nil {
		return diag.FromErr(err)
	}

	mode := r.Get(keyDestConMode).(string)
	enabled := r.Get(keyDestEnabled).(bool)
	var config []segment.DestinationConfig
	if d := decodeDestinationConfig(meta.workspace, srcName, destName, r.Get("config"), &config); d != nil {
		return *d
	}

	if _, err := client.CreateDestination(srcName, destName, mode, enabled, config); err != nil {
		return diag.FromErr(err)
	}

	r.SetId(id)

	return resourceSegmentDestinationRead(ctx, r, m)
}

func resourceSegmentDestinationDelete(_ context.Context, r *schema.ResourceData, m interface{}) diag.Diagnostics {
	meta := m.(ProviderMetadata)
	client := meta.client
	srcName := r.Get(keySource).(string)
	destName := r.Get(keyDestName).(string)

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

		config.Name = k
		if d := validateConfigValue(config); d != nil {
			return d
		}

		// TODO: Version will error on update to the Config API, remove when it's fixed
		if k == "version" {
			continue
		}

		config.Name = fmt.Sprintf("workspaces/%s/sources/%s/destinations/%s/config/%s", workspace, srcName, destName, k)

		*dst = append(*dst, config)
	}

	return nil
}

func validateConfigValue(config segment.DestinationConfig) *diag.Diagnostics {
	switch config.Type {
	case "string", "password":
		if _, ok := config.Value.(string); !ok {
			return configTypeError(config.Name, config.Type, config.Value)
		}
	case "number":
		if _, ok := config.Value.(float32); !ok {
			return configTypeError(config.Name, config.Type, config.Value)
		}
	case "boolean":
		if _, ok := config.Value.(bool); !ok {
			return configTypeError(config.Name, config.Type, config.Value)
		}
	case "select":
		if _, ok := config.Value.(string); !ok {
			return configTypeError(config.Name, config.Type, config.Value)
		}
	case "mixed":
		if _, ok := config.Value.([]interface{}); !ok {
			return configTypeError(config.Name, config.Type, config.Value)
		}
	default:
		return nil
	}

	return nil
}

func validateDestinationConfig(i interface{}, _ cty.Path) diag.Diagnostics {
	var c []segment.DestinationConfig
	if d := decodeDestinationConfig("test", "test", "test", i, &c); d != nil {
		return *d
	}
	return nil
}

func configTypeError(name string, typ string, value interface{}) *diag.Diagnostics {
	d := diag.Errorf("Unexpected config value for %s of expected type %s: %v", name, typ, value)
	return &d
}

// Misc Helpers

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
