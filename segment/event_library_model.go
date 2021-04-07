package segment

import (
	"github.com/uswitch/segment-config-go/segment"
)

type Property struct {
	Type        []string `json:"type"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Required    bool     `json:"required"`
}

type Event struct {
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Properties  []Property `json:"properties"`
}

type EventLibrary struct {
	Events []Event `json:"events"`
}

func (evtLib *EventLibrary) convertToSegmentEvents() []segment.Event {
	segmentEvents := make([]segment.Event, len(evtLib.Events))

	for i, event := range evtLib.Events {
		properties := segment.Properties{}

		eventProperties := map[string]segment.Property{}
		for _, prop := range event.Properties {
			eventProperties[prop.Name] = segment.Property{
				Description: prop.Description,
				Type:        prop.Type,
			}
			if prop.Required {
				properties.Required = append(properties.Required, prop.Name)
			}
		}
		properties.Properties = eventProperties

		rules := segment.Rules{
			Schema: "http://json-schema.org/draft-07/schema#",
			Type:   "object",
			Properties: segment.RuleProperties{
				Context:    segment.Properties{},
				Traits:     segment.Properties{},
				Properties: properties,
			},
		}

		event := segment.Event{
			Name:        event.Name,
			Description: event.Description,
			Rules:       rules,
		}
		segmentEvents[i] = event
	}

	return segmentEvents
}
