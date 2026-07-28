package lariv

import "encoding/json"

// GrapesJSBlock is the props object passed as the second argument to
// GrapesJS BlockManager.add(id, props). The registry entry Key is the block id.
//
// Content, Category, and OnClick use json.RawMessage so plugins can supply either
// a JSON string or an object (or, for OnClick, a boolean or JS function-body string).
type GrapesJSBlock struct {
	Label      string          `json:"label"`
	Content    json.RawMessage `json:"content"`
	Media      string          `json:"media,omitempty"`
	Category   json.RawMessage `json:"category,omitempty"`
	Attributes map[string]any  `json:"attributes,omitempty"`
	Activate   bool            `json:"activate,omitempty"`
	Select     bool            `json:"select,omitempty"`
	Disable    bool            `json:"disable,omitempty"`
	OnClick    json.RawMessage `json:"onClick,omitempty"`
}

// GrapesJSBlockContentString marshals an HTML (or HTML+style) string for Content.
func GrapesJSBlockContentString(html string) json.RawMessage {
	b, err := json.Marshal(html)
	if err != nil {
		return json.RawMessage(`""`)
	}
	return b
}

// GrapesJSBlockContentJSON marshals a component-definition object for Content.
func GrapesJSBlockContentJSON(v any) (json.RawMessage, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	return b, nil
}

// GrapesJSBlockCategoryString marshals a category label string.
func GrapesJSBlockCategoryString(label string) json.RawMessage {
	b, err := json.Marshal(label)
	if err != nil {
		return json.RawMessage(`""`)
	}
	return b
}

// GrapesJSBlockOnClickBool marshals a boolean onClick value.
func GrapesJSBlockOnClickBool(v bool) json.RawMessage {
	b, err := json.Marshal(v)
	if err != nil {
		return json.RawMessage(`false`)
	}
	return b
}

// GrapesJSBlockOnClickJS marshals a trusted JS function body for onClick.
// The body receives parameters block and editor when invoked in the builder.
func GrapesJSBlockOnClickJS(body string) json.RawMessage {
	b, err := json.Marshal(body)
	if err != nil {
		return json.RawMessage(`""`)
	}
	return b
}
