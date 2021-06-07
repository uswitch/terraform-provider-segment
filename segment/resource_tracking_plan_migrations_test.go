package segment_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/uswitch/terraform-provider-segment/segment"
)

func TestResourceInstanceStateUpgradeV1(t *testing.T) {
	tests := map[string]struct {
		v1Config map[string]interface{}
		v2Config map[string]interface{}
	}{
		"no import_from defined": {
			v1Config: map[string]interface{}{"import_from": nil},
			v2Config: map[string]interface{}{"import_from": nil},
		},
		"empty import_from": {
			v1Config: map[string]interface{}{"import_from": []interface{}{}},
			v2Config: map[string]interface{}{"import_from": "[]"},
		},
		"single lib": {
			v1Config: map[string]interface{}{
				"import_from": []interface{}{
					"{\"global\":{\"$schema\":\"http://json-schema.org/draft-07/schema#\",\"type\":\"object\",\"properties\":{\"context\":{},\"properties\":{},\"traits\":{}}},\"events\":[{\"name\":\"Form Submitted\",\"description\":\"fired when a user submits a form.\",\"rules\":{\"$schema\":\"http://json-schema.org/draft-07/schema#\",\"type\":\"object\",\"properties\":{\"context\":{},\"properties\":{\"required\":[\"form_id\"],\"type\":\"object\",\"properties\":{\"form_id\":{\"description\":\"the ID of the form\",\"type\":[\"string\"]},\"name\":{\"description\":\"the name of the form\",\"type\":[\"string\",\"null\"]},\"form_orientation\":{\"description\":\"orientation of the form\",\"type\":[\"integer\"]}}},\"traits\":{}},\"required\":[\"properties\"]},\"version\":1},{\"name\":\"Button Pressed\",\"description\":\"fired when a user submits a form.\",\"rules\":{\"$schema\":\"http://json-schema.org/draft-07/schema#\",\"type\":\"object\",\"properties\":{\"context\":{},\"properties\":{\"required\":[\"form_id\"],\"type\":\"object\",\"properties\":{\"form_id\":{\"description\":\"the ID of the form\",\"type\":[\"string\"]},\"name\":{\"description\":\"the name of the form\",\"type\":[\"string\",\"null\"]},\"form_orientation\":{\"description\":\"orientation of the form\",\"type\":[\"integer\"]}}},\"traits\":{}},\"required\":[\"properties\"]},\"version\":1}],\"group\":{\"$schema\":\"http://json-schema.org/draft-07/schema#\",\"type\":\"object\",\"properties\":{\"context\":{},\"properties\":{},\"traits\":{}}},\"identify\":{\"$schema\":\"http://json-schema.org/draft-07/schema#\",\"type\":\"object\",\"properties\":{\"context\":{},\"properties\":{},\"traits\":{}}}}",
				}},
			v2Config: map[string]interface{}{
				"import_from": "[{\"global\":{\"$schema\":\"http://json-schema.org/draft-07/schema#\",\"type\":\"object\",\"properties\":{\"context\":{},\"properties\":{},\"traits\":{}}},\"events\":[{\"name\":\"Form Submitted\",\"description\":\"fired when a user submits a form.\",\"rules\":{\"$schema\":\"http://json-schema.org/draft-07/schema#\",\"type\":\"object\",\"properties\":{\"context\":{},\"properties\":{\"required\":[\"form_id\"],\"type\":\"object\",\"properties\":{\"form_id\":{\"description\":\"the ID of the form\",\"type\":[\"string\"]},\"name\":{\"description\":\"the name of the form\",\"type\":[\"string\",\"null\"]},\"form_orientation\":{\"description\":\"orientation of the form\",\"type\":[\"integer\"]}}},\"traits\":{}},\"required\":[\"properties\"]},\"version\":1},{\"name\":\"Button Pressed\",\"description\":\"fired when a user submits a form.\",\"rules\":{\"$schema\":\"http://json-schema.org/draft-07/schema#\",\"type\":\"object\",\"properties\":{\"context\":{},\"properties\":{\"required\":[\"form_id\"],\"type\":\"object\",\"properties\":{\"form_id\":{\"description\":\"the ID of the form\",\"type\":[\"string\"]},\"name\":{\"description\":\"the name of the form\",\"type\":[\"string\",\"null\"]},\"form_orientation\":{\"description\":\"orientation of the form\",\"type\":[\"integer\"]}}},\"traits\":{}},\"required\":[\"properties\"]},\"version\":1}],\"group\":{\"$schema\":\"http://json-schema.org/draft-07/schema#\",\"type\":\"object\",\"properties\":{\"context\":{},\"properties\":{},\"traits\":{}}},\"identify\":{\"$schema\":\"http://json-schema.org/draft-07/schema#\",\"type\":\"object\",\"properties\":{\"context\":{},\"properties\":{},\"traits\":{}}}}]",
			}},
		"multiple libs": {
			v1Config: map[string]interface{}{
				"import_from": []interface{}{
					"{\"global\":{\"$schema\":\"http://json-schema.org/draft-07/schema#\",\"type\":\"object\",\"properties\":{\"context\":{},\"properties\":{},\"traits\":{}}},\"events\":[{\"name\":\"Form Submitted\",\"description\":\"fired when a user submits a form.\",\"rules\":{\"$schema\":\"http://json-schema.org/draft-07/schema#\",\"type\":\"object\",\"properties\":{\"context\":{},\"properties\":{\"required\":[\"form_id\"],\"type\":\"object\",\"properties\":{\"form_id\":{\"description\":\"the ID of the form\",\"type\":[\"string\"]},\"name\":{\"description\":\"the name of the form\",\"type\":[\"string\",\"null\"]},\"form_orientation\":{\"description\":\"orientation of the form\",\"type\":[\"integer\"]}}},\"traits\":{}},\"required\":[\"properties\"]},\"version\":1},{\"name\":\"Button Pressed\",\"description\":\"fired when a user submits a form.\",\"rules\":{\"$schema\":\"http://json-schema.org/draft-07/schema#\",\"type\":\"object\",\"properties\":{\"context\":{},\"properties\":{\"required\":[\"form_id\"],\"type\":\"object\",\"properties\":{\"form_id\":{\"description\":\"the ID of the form\",\"type\":[\"string\"]},\"name\":{\"description\":\"the name of the form\",\"type\":[\"string\",\"null\"]},\"form_orientation\":{\"description\":\"orientation of the form\",\"type\":[\"integer\"]}}},\"traits\":{}},\"required\":[\"properties\"]},\"version\":1}],\"group\":{\"$schema\":\"http://json-schema.org/draft-07/schema#\",\"type\":\"object\",\"properties\":{\"context\":{},\"properties\":{},\"traits\":{}}},\"identify\":{\"$schema\":\"http://json-schema.org/draft-07/schema#\",\"type\":\"object\",\"properties\":{\"context\":{},\"properties\":{},\"traits\":{}}}}",
					"{\"global\":{\"$schema\":\"http://json-schema.org/draft-07/schema#\",\"type\":\"object\",\"properties\":{\"context\":{},\"properties\":{},\"traits\":{}}},\"events\":[{\"name\":\"Broadband Form Submitted\",\"description\":\"fired when a user submits a form.\",\"rules\":{\"$schema\":\"http://json-schema.org/draft-07/schema#\",\"type\":\"object\",\"properties\":{\"context\":{},\"properties\":{\"required\":[\"form_id\"],\"type\":\"object\",\"properties\":{\"form_id\":{\"description\":\"the ID of the form\",\"type\":[\"string\"]},\"name\":{\"description\":\"the name of the form\",\"type\":[\"string\",\"null\"]},\"form_orientation\":{\"description\":\"orientation of the form\",\"type\":[\"integer\"]}}},\"traits\":{}},\"required\":[\"properties\"]},\"version\":1},{\"name\":\"Broadband Button Pressed\",\"description\":\"fired when a user submits a form.\",\"rules\":{\"$schema\":\"http://json-schema.org/draft-07/schema#\",\"type\":\"object\",\"properties\":{\"context\":{},\"properties\":{\"required\":[\"form_id\"],\"type\":\"object\",\"properties\":{\"form_id\":{\"description\":\"the ID of the form\",\"type\":[\"string\"]},\"name\":{\"description\":\"the name of the form\",\"type\":[\"string\",\"null\"]},\"form_orientation\":{\"description\":\"orientation of the form\",\"type\":[\"integer\"]}}},\"traits\":{}},\"required\":[\"properties\"]},\"version\":1}],\"group\":{\"$schema\":\"http://json-schema.org/draft-07/schema#\",\"type\":\"object\",\"properties\":{\"context\":{},\"properties\":{},\"traits\":{}}},\"identify\":{\"$schema\":\"http://json-schema.org/draft-07/schema#\",\"type\":\"object\",\"properties\":{\"context\":{},\"properties\":{},\"traits\":{}}}}",
				},
			},
			v2Config: map[string]interface{}{
				"import_from": "[{\"global\":{\"$schema\":\"http://json-schema.org/draft-07/schema#\",\"type\":\"object\",\"properties\":{\"context\":{},\"properties\":{},\"traits\":{}}},\"events\":[{\"name\":\"Form Submitted\",\"description\":\"fired when a user submits a form.\",\"rules\":{\"$schema\":\"http://json-schema.org/draft-07/schema#\",\"type\":\"object\",\"properties\":{\"context\":{},\"properties\":{\"required\":[\"form_id\"],\"type\":\"object\",\"properties\":{\"form_id\":{\"description\":\"the ID of the form\",\"type\":[\"string\"]},\"name\":{\"description\":\"the name of the form\",\"type\":[\"string\",\"null\"]},\"form_orientation\":{\"description\":\"orientation of the form\",\"type\":[\"integer\"]}}},\"traits\":{}},\"required\":[\"properties\"]},\"version\":1},{\"name\":\"Button Pressed\",\"description\":\"fired when a user submits a form.\",\"rules\":{\"$schema\":\"http://json-schema.org/draft-07/schema#\",\"type\":\"object\",\"properties\":{\"context\":{},\"properties\":{\"required\":[\"form_id\"],\"type\":\"object\",\"properties\":{\"form_id\":{\"description\":\"the ID of the form\",\"type\":[\"string\"]},\"name\":{\"description\":\"the name of the form\",\"type\":[\"string\",\"null\"]},\"form_orientation\":{\"description\":\"orientation of the form\",\"type\":[\"integer\"]}}},\"traits\":{}},\"required\":[\"properties\"]},\"version\":1}],\"group\":{\"$schema\":\"http://json-schema.org/draft-07/schema#\",\"type\":\"object\",\"properties\":{\"context\":{},\"properties\":{},\"traits\":{}}},\"identify\":{\"$schema\":\"http://json-schema.org/draft-07/schema#\",\"type\":\"object\",\"properties\":{\"context\":{},\"properties\":{},\"traits\":{}}}}, {\"global\":{\"$schema\":\"http://json-schema.org/draft-07/schema#\",\"type\":\"object\",\"properties\":{\"context\":{},\"properties\":{},\"traits\":{}}},\"events\":[{\"name\":\"Broadband Form Submitted\",\"description\":\"fired when a user submits a form.\",\"rules\":{\"$schema\":\"http://json-schema.org/draft-07/schema#\",\"type\":\"object\",\"properties\":{\"context\":{},\"properties\":{\"required\":[\"form_id\"],\"type\":\"object\",\"properties\":{\"form_id\":{\"description\":\"the ID of the form\",\"type\":[\"string\"]},\"name\":{\"description\":\"the name of the form\",\"type\":[\"string\",\"null\"]},\"form_orientation\":{\"description\":\"orientation of the form\",\"type\":[\"integer\"]}}},\"traits\":{}},\"required\":[\"properties\"]},\"version\":1},{\"name\":\"Broadband Button Pressed\",\"description\":\"fired when a user submits a form.\",\"rules\":{\"$schema\":\"http://json-schema.org/draft-07/schema#\",\"type\":\"object\",\"properties\":{\"context\":{},\"properties\":{\"required\":[\"form_id\"],\"type\":\"object\",\"properties\":{\"form_id\":{\"description\":\"the ID of the form\",\"type\":[\"string\"]},\"name\":{\"description\":\"the name of the form\",\"type\":[\"string\",\"null\"]},\"form_orientation\":{\"description\":\"orientation of the form\",\"type\":[\"integer\"]}}},\"traits\":{}},\"required\":[\"properties\"]},\"version\":1}],\"group\":{\"$schema\":\"http://json-schema.org/draft-07/schema#\",\"type\":\"object\",\"properties\":{\"context\":{},\"properties\":{},\"traits\":{}}},\"identify\":{\"$schema\":\"http://json-schema.org/draft-07/schema#\",\"type\":\"object\",\"properties\":{\"context\":{},\"properties\":{},\"traits\":{}}}}]",
			},
		},
	}

	for testName, test := range tests {
		t.Run(testName, func(t *testing.T) {
			result, err := segment.TpV1V2Upgrader().Upgrade(context.Background(), test.v1Config, nil)
			assert.NoError(t, err)
			assert.NotNil(t, result)

			switch result["import_from"].(type) {
			case string:
				assert.JSONEq(t, test.v2Config["import_from"].(string), result["import_from"].(string))
			default:
				assert.Equal(t, test.v2Config["import_from"], result["import_from"])
			}
		})
	}
}
