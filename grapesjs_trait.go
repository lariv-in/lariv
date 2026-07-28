package lariv

import "encoding/json"

// GrapesJSTrait is the props object passed as the second argument to
// GrapesJS Traits.addType(id, props). The registry entry Key is the trait type id.
//
// TemplateInput, CreateInput, CreateLabel, OnEvent, and OnUpdate use json.RawMessage
// so plugins can supply JSON strings/objects or trusted JS function-body strings
// (revived in the builder).
type GrapesJSTrait struct {
	NoLabel       bool            `json:"noLabel,omitempty"`
	EventCapture  []string        `json:"eventCapture,omitempty"`
	TemplateInput json.RawMessage `json:"templateInput,omitempty"`
	CreateInput   json.RawMessage `json:"createInput,omitempty"`
	CreateLabel   json.RawMessage `json:"createLabel,omitempty"`
	OnEvent       json.RawMessage `json:"onEvent,omitempty"`
	OnUpdate      json.RawMessage `json:"onUpdate,omitempty"`
}

// GrapesJSTraitCreateInputJS marshals a trusted JS function body for createInput.
// The body receives a destructured props object ({ trait }) when invoked.
func GrapesJSTraitCreateInputJS(body string) json.RawMessage {
	return marshalJSBody(body)
}

// GrapesJSTraitCreateLabelJS marshals a trusted JS function body for createLabel.
func GrapesJSTraitCreateLabelJS(body string) json.RawMessage {
	return marshalJSBody(body)
}

// GrapesJSTraitOnEventJS marshals a trusted JS function body for onEvent.
// The body receives a destructured props object ({ elInput, component, event }).
func GrapesJSTraitOnEventJS(body string) json.RawMessage {
	return marshalJSBody(body)
}

// GrapesJSTraitOnUpdateJS marshals a trusted JS function body for onUpdate.
// The body receives a destructured props object ({ elInput, component }).
func GrapesJSTraitOnUpdateJS(body string) json.RawMessage {
	return marshalJSBody(body)
}

// GrapesJSTraitTemplateInputString marshals a templateInput HTML string.
func GrapesJSTraitTemplateInputString(html string) json.RawMessage {
	b, err := json.Marshal(html)
	if err != nil {
		return json.RawMessage(`""`)
	}
	return b
}

// GrapesJSTraitTemplateInputJS marshals a trusted JS function body for templateInput.
func GrapesJSTraitTemplateInputJS(body string) json.RawMessage {
	return marshalJSBody(body)
}

func marshalJSBody(body string) json.RawMessage {
	b, err := json.Marshal(body)
	if err != nil {
		return json.RawMessage(`""`)
	}
	return b
}
