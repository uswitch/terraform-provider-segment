package segment

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"reflect"
	"strings"

	"github.com/ajbosco/segment-config-go/segment"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/uswitch/terraform-provider-segment/segment/internal/utils"
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
		Description: "A destination connection on Segment. More information on destinations and how to use them can be found in the [Segment Destinations documentation](https://segment.com/docs/connections/destinations/).",
		Schema: map[string]*schema.Schema{
			keyDestSource: {
				Description: "The Segment source name this destination is connecting to.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			keyDestName: {
				Description: "The name of the destination.",
				Type:        schema.TypeString,
				Required:    true,
			},
			keyDestEnabled: {
				Description: "Whether the destination is enabled and can receive events.",
				Type:        schema.TypeBool,
				Required:    true,
			},
			keyDestParent: {
				Description: "The source the destination is associated with. *(Set by Segment)*.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			keyDestDisplayName: {
				Description: "The display name of the destination in the Segment UI. *(Set by Segment)*.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			keyDestCreateTime: {
				Description: "The time the destination was created. *(Set by Segment)*.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			keyDestUpdateTime: {
				Description: "The time the destination was last updated. *(Set by Segment)*.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			keyDestConMode: {
				Description:  "The connection type of the destination. Available values are: `UNSPECIFIED`, `CLOUD`, `DEVICE`.",
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringInSlice([]string{"UNSPECIFIED", "CLOUD", "DEVICE"}, false),
			},
			keyDestConfig: {
				Description: "The configuration of the destination. This varies according to the destination. The specific fields can be retrieved by making a request to the [Get Destination](https://reference.segmentapis.com/#94aed763-b2bd-4ee6-8b5b-b6d39aacba21) endpoint.",
				Type:        schema.TypeMap,
				Required:    true,
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

func resourceSegmentDestinationRead(c context.Context, r *schema.ResourceData, m interface{}) diag.Diagnostics {
	meta := m.(ProviderMetadata)
	client := meta.Client
	srcName, dstName := destinationIdToSourceAndDest(r.Id())

	d, err := client.GetDestination(srcName, dstName)
	if err != nil {
		return diag.FromErr(err)
	}

	config := map[string]interface{}{}
	utils.CatchFirst(
		func() error { return encodeDestinationConfig(d, &config) },
		func() error { return r.Set(keyDestSource, srcName) },
		func() error { return r.Set(keyDestName, utils.PathToName(d.Name)) },
		func() error { return r.Set(keyDestEnabled, d.Enabled) },
		func() error { return r.Set(keyDestParent, d.Parent) },
		func() error { return r.Set(keyDestDisplayName, d.DisplayName) },
		func() error { return r.Set(keyDestConMode, d.ConnectionMode) },
		func() error { return r.Set(keyDestConfig, config) },
		func() error { return r.Set(keyDestCreateTime, d.CreateTime.String()) },
		func() error { return r.Set(keyDestUpdateTime, d.UpdateTime.String()) },
	)

	return nil
}

func resourceSegmentDestinationUpdate(ctx context.Context, r *schema.ResourceData, m interface{}) diag.Diagnostics {
	meta := m.(ProviderMetadata)
	client := meta.Client
	srcName := r.Get(keyDestSource).(string)
	destName := r.Get(keyDestName).(string)
	enabled := r.Get(keyDestEnabled).(bool)
	rawConfig := r.Get(keyDestConfig).(map[string]interface{})

	config := []segment.DestinationConfig{}
	if d := decodeDestinationConfig(meta.Workspace, srcName, destName, rawConfig, &config, meta.IsDestinationConfigPropSupported); d != nil {
		return d
	}

	if _, err := client.UpdateDestination(srcName, destName, enabled, config); err != nil {
		return diag.FromErr(err)
	}

	return resourceSegmentDestinationRead(ctx, r, m)
}

func resourceSegmentDestinationCreate(ctx context.Context, r *schema.ResourceData, m interface{}) diag.Diagnostics {
	meta := m.(ProviderMetadata)
	client := meta.Client
	srcName := r.Get(keyDestSource).(string)
	destName := r.Get(keyDestName).(string)
	id := destinationResourceId(srcName, destName)

	mode := r.Get(keyDestConMode).(string)
	enabled := r.Get(keyDestEnabled).(bool)
	var config []segment.DestinationConfig
	if d := decodeDestinationConfig(meta.Workspace, srcName, destName, r.Get("config"), &config, meta.IsDestinationConfigPropSupported); d != nil {
		return d
	}

	log.Printf("[INFO] Creating destination %s for %s", destName, srcName)
	if _, err := client.CreateDestination(srcName, destName, mode, enabled, config); err != nil {
		return diag.FromErr(err)
	}

	r.SetId(id)

	return resourceSegmentDestinationRead(ctx, r, m)
}

func resourceSegmentDestinationDelete(_ context.Context, r *schema.ResourceData, m interface{}) diag.Diagnostics {
	meta := m.(ProviderMetadata)
	client := meta.Client
	srcName, destName := destinationIdToSourceAndDest(r.Id())

	err := client.DeleteDestination(srcName, destName)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

// Decoders

func encodeDestinationConfig(destination segment.Destination, encoded *map[string]interface{}) error {
	if encoded == nil {
		return errors.New("destination config encoded map cannot be nil")
	}

	for _, config := range destination.Configs {
		c, err := json.Marshal(map[string]interface{}{
			"type":  config.Type,
			"value": config.Value,
		})
		if err != nil {
			utils.DiagFromErrPtr(err)
		}

		(*encoded)[utils.PathToName(config.Name)] = string(c)
	}

	return nil
}

func decodeDestinationConfig(workspace string, srcName string, destName string, rawConfig interface{}, dst *[]segment.DestinationConfig, isPropAllowed func(d string, k string) bool) (diags diag.Diagnostics) {
	defer func() {
		if r := recover(); r != nil {
			diags = diag.FromErr(fmt.Errorf("failed to decode destination config: %w", r.(error)))
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

		if !isPropAllowed(destName, k) {
			continue
		}

		config.Name = fmt.Sprintf("workspaces/%s/sources/%s/destinations/%s/config/%s", workspace, srcName, destName, k)

		*dst = append(*dst, config)
	}

	return nil
}

func validateConfigValue(config segment.DestinationConfig) diag.Diagnostics {
	switch config.Type {
	case "string", "password":
		if _, ok := config.Value.(string); !ok {
			return configTypeError(config.Name, config.Type, config.Value)
		}
	case "number":
		floatType := reflect.TypeOf(float64(0))
		v := reflect.Indirect(reflect.ValueOf(config.Value))
		if !v.Type().ConvertibleTo(floatType) {
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
	if d := decodeDestinationConfig("test", "test", "test", i, &c, func(d string, k string) bool { return false }); d != nil {
		return d
	}
	return nil
}

func configTypeError(name string, typ string, value interface{}) diag.Diagnostics {
	d := diag.Errorf("Unexpected config value for %s of expected type %s: %v", name, typ, value)
	return d
}

// Misc Helpers

func destinationIdToSourceAndDest(id string) (string, string) {
	parts := strings.Split(id, "/")

	if len(parts) > 2 {
		panic("Invalid destination id: " + id)
	}

	return parts[0], parts[1]
}

func destinationResourceId(src string, dst string) string {
	return src + "/" + dst
}
