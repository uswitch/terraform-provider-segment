package segment

import (
	"context"
	"log"

	"github.com/ajbosco/segment-config-go/segment"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/uswitch/terraform-provider-segment/segment/internal/utils"
)

func resourceSegmentDestinationFilter() *schema.Resource {
	return &schema.Resource{
		Schema:        map[string]*schema.Schema{},
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
	if err := r.Set(keyDestName, utils.PathToName(d.Name)); err != nil {
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

func resourceSegmentDestinationFilterUpdate(ctx context.Context, r *schema.ResourceData, m interface{}) diag.Diagnostics {
	meta := m.(ProviderMetadata)
	client := meta.client
	srcName := r.Get(keyDestSource).(string)
	destName := r.Get(keyDestName).(string)
	enabled := r.Get(keyDestEnabled).(bool)
	rawConfig := r.Get(keyDestConfig).(map[string]interface{})

	config := []segment.DestinationConfig{}
	if d := decodeDestinationConfig(meta.workspace, srcName, destName, rawConfig, &config, meta.isDestinationConfigPropSupported); d != nil {
		return *d
	}

	if _, err := client.UpdateDestination(srcName, destName, enabled, config); err != nil {
		return diag.FromErr(err)
	}

	return resourceSegmentDestinationFilterRead(ctx, r, m)
}

func resourceSegmentDestinationFilterCreate(ctx context.Context, r *schema.ResourceData, m interface{}) diag.Diagnostics {
	meta := m.(ProviderMetadata)
	client := meta.client
	srcName := r.Get(keyDestSource).(string)
	destName := r.Get(keyDestName).(string)
	id := destinationResourceId(srcName, destName)

	mode := r.Get(keyDestConMode).(string)
	enabled := r.Get(keyDestEnabled).(bool)
	var config []segment.DestinationConfig
	if d := decodeDestinationConfig(meta.workspace, srcName, destName, r.Get("config"), &config, meta.isDestinationConfigPropSupported); d != nil {
		return *d
	}

	log.Printf("[INFO] Creating destination %s for %s", destName, srcName)
	if _, err := client.CreateDestination(srcName, destName, mode, enabled, config); err != nil {
		return diag.FromErr(err)
	}

	r.SetId(id)

	return resourceSegmentDestinationFilterRead(ctx, r, m)
}

func resourceSegmentDestinationFilterDelete(_ context.Context, r *schema.ResourceData, m interface{}) diag.Diagnostics {
	meta := m.(ProviderMetadata)
	client := meta.client
	srcName, destName := idToSourceAndDest(r)

	err := client.DeleteDestination(srcName, destName)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

// Decoders

// Misc Helpers
