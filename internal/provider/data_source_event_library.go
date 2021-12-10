package provider

import (
	"context"
	"encoding/json"
	"strconv"

	"github.com/ajbosco/segment-config-go/segment"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/uswitch/terraform-provider-segment/internal/hashcode"
	"github.com/uswitch/terraform-provider-segment/internal/utils"
)

func dataSourceEventLibrary() *schema.Resource {
	return &schema.Resource{
		Description: "A data source representing an event library which is a set of common tracking plan rules which can be reused by the provider to define common attributes in multiple tracking plans.",
		ReadContext: dataSourceEventLibraryRead,
		Schema: map[string]*schema.Schema{
			"rules_json_file": {
				Description:      "The location of the JSON schema file containing the event library rules. The `file()` terraform function should be used and an absolute or relative path can be passed.",
				Type:             schema.TypeString,
				Optional:         true,
				ValidateFunc:     validation.StringIsJSON,
				DiffSuppressFunc: utils.DiffRulesJSONState,
			},
			"json": {
				Description: "A string representation of the json associated with an event library.",
				Type:        schema.TypeString,
				Computed:    true,
			},
		},
	}
}

func dataSourceEventLibraryRead(_ context.Context, d *schema.ResourceData, _ interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	// Read tracking plan rules
	var inputRules segment.RuleSet
	if inputJSON, ok := d.GetOk("rules_json_file"); ok {
		if err := json.Unmarshal([]byte(inputJSON.(string)), &inputRules); err != nil {
			return diag.FromErr(err)
		}
	}
	// Convert Rules to JSON
	outputJSON, err := json.Marshal(inputRules)
	if err != nil {
		return diag.FromErr(err)
	}
	jsonString := string(outputJSON)

	d.Set("json", jsonString)
	d.SetId(strconv.Itoa(hashcode.String(jsonString)))

	return diags
}
