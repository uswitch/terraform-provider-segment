package segment

import (
	"context"
	"encoding/json"
	"strconv"

	"github.com/ajbosco/segment-config-go/segment"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/uswitch/terraform-provider-segment/segment/internal/hashcode"
	"github.com/uswitch/terraform-provider-segment/segment/internal/utils"
)

func dataSourceEventLibrary() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceEventLibraryRead,
		Schema: map[string]*schema.Schema{
			"rules_json_file": {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateFunc:     validation.StringIsJSON,
				DiffSuppressFunc: utils.DiffRulesJSONState,
			},
			"json": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceEventLibraryRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
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
