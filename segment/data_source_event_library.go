package segment

import (
	"encoding/json"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/uswitch/terraform-provider-segment/segment/helpers/hashcode"
)

func dataSourceEventLibrary() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceEventLibraryRead,
		Schema: map[string]*schema.Schema{
			"json": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"event": {
				Type:     schema.TypeList,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"description": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"property": {
							Type:     schema.TypeList,
							Required: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"name": {
										Type:     schema.TypeString,
										Required: true,
									},
									"description": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"type": {
										Type:     schema.TypeList,
										Required: true,
										Elem: &schema.Schema{
											Type: schema.TypeString,
										},
									},
									"required": {
										Type:     schema.TypeBool,
										Required: true,
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func dataSourceEventLibraryRead(d *schema.ResourceData, m interface{}) error {
	eventLib := EventLibrary{
		Events: []Event{},
	}

	if events, ok := d.GetOk("event"); ok {
		var eventsIntf = events.([]interface{})
		evts := make([]Event, len(eventsIntf))

		for i, eventI := range eventsIntf {
			var event = eventI.(map[string]interface{})

			var propertiesIntf = event["property"].([]interface{})
			props := make([]Property, len(propertiesIntf))

			for j, propertyIntf := range propertiesIntf {
				var property = propertyIntf.(map[string]interface{})

				propDesc := ""
				if pDesc, ok := property["description"]; ok {
					propDesc = pDesc.(string)
				}

				var propTypesIntf = property["type"].([]interface{})
				propTypes := make([]string, len(propTypesIntf))

				for p, propTypeIntf := range propTypesIntf {
					propTypes[p] = propTypeIntf.(string)
				}

				props[j] = Property{
					Name:        property["name"].(string),
					Description: propDesc,
					Type:        propTypes,
					Required:    property["required"].(bool),
				}
			}

			eventDesc := ""
			if eDesc, ok := event["description"]; ok {
				eventDesc = eDesc.(string)
			}

			evts[i] = Event{
				Name:        event["name"].(string),
				Description: eventDesc,
				Properties:  props,
			}
		}

		eventLib.Events = evts
	}

	jsonDoc, err := json.MarshalIndent(eventLib, "", "  ")
	if err != nil {
		return err
	}
	jsonString := string(jsonDoc)

	d.Set("json", jsonString)
	d.Set("Id", hashcode.String(jsonString))

	return nil
}
