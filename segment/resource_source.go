package segment

import (
	"context"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
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
		},
		CreateContext: resourceSegmentSourceCreate,
		ReadContext:   resourceSegmentSourceRead,
		DeleteContext: resourceSegmentSourceDelete,
		// UpdateContext: resourceSegmentSourceUpdate,
		// Importer: &schema.ResourceImporter{
		// 	State: resourceSegmentSourceImport,
		// },
	}
}

func resourceSegmentSourceRead(ctx context.Context, r *schema.ResourceData, m interface{}) diag.Diagnostics {
	meta := m.(SegmentMetadata)
	id := r.Id()

	s, err := meta.client.GetSource(id)
	if err != nil {
		return diag.FromErr(err)
	}

	r.Set("catalog_name", s.CatalogName)

	return nil
}

func resourceSegmentSourceCreate(ctx context.Context, r *schema.ResourceData, m interface{}) diag.Diagnostics {
	meta := m.(SegmentMetadata)
	srcName := r.Get("source_name").(string)
	catName := r.Get("catalog_name").(string)

	_, err := meta.client.CreateSource(srcName, catName)
	if err != nil {
		log.Printf("[ERROR] Problem creating source %s \n", err)
		return diag.FromErr(err)
	}

	r.SetId(srcName)

	return resourceSegmentSourceRead(ctx, r, m)
}

func resourceSegmentSourceUpdate(ctx context.Context, r *schema.ResourceData, m interface{}) diag.Diagnostics {
	meta := m.(SegmentMetadata)
	srcName := r.Get("source_name").(string)
	catName := r.Get("catalog_name").(string)

	source, err := meta.client.CreateSource(srcName, catName)
	if err != nil {
		log.Printf("[ERROR] Problem creating source %s \n", err)
		return diag.FromErr(err)
	}

	r.SetId(source.Name)

	return resourceSegmentSourceRead(ctx, r, m)
}

func resourceSegmentSourceDelete(ctx context.Context, r *schema.ResourceData, m interface{}) diag.Diagnostics {
	meta := m.(SegmentMetadata)
	id := r.Id()

	err := meta.client.DeleteSource(id)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}
