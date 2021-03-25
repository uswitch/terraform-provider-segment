package segment

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
