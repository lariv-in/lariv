package lariv

import "encoding/json"

// GrapesJSComponent is the props object passed as the second argument to
// GrapesJS DomComponents.addType(id, props). The registry entry Key is the type id.
//
// IsComponent, Model, and View use json.RawMessage so plugins can supply JSON
// objects/bools or trusted JS function-body strings (revived in the builder).
type GrapesJSComponent struct {
	Extend      string          `json:"extend,omitempty"`
	IsComponent json.RawMessage `json:"isComponent,omitempty"`
	Model       json.RawMessage `json:"model,omitempty"`
	View        json.RawMessage `json:"view,omitempty"`
}

// GrapesJSComponentIsComponentJS marshals a trusted JS function body for isComponent.
// The body receives parameter el when invoked in the builder.
func GrapesJSComponentIsComponentJS(body string) json.RawMessage {
	b, err := json.Marshal(body)
	if err != nil {
		return json.RawMessage(`""`)
	}
	return b
}

// GrapesJSComponentIsComponentBool marshals a boolean isComponent value.
func GrapesJSComponentIsComponentBool(v bool) json.RawMessage {
	b, err := json.Marshal(v)
	if err != nil {
		return json.RawMessage(`false`)
	}
	return b
}

// GrapesJSComponentIsComponentTag returns an isComponent JS body that matches a tag name.
func GrapesJSComponentIsComponentTag(tag string) json.RawMessage {
	escaped, err := json.Marshal(tag)
	if err != nil {
		return GrapesJSComponentIsComponentJS(`return false;`)
	}
	return GrapesJSComponentIsComponentJS(
		`return el && el.tagName && el.tagName.toLowerCase() === ` + string(escaped) + `.toLowerCase();`,
	)
}

// GrapesJSComponentModelJSON marshals a model definition object.
func GrapesJSComponentModelJSON(v any) (json.RawMessage, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	return b, nil
}

// GrapesJSComponentMustModelJSON marshals a model definition or panics.
func GrapesJSComponentMustModelJSON(v any) json.RawMessage {
	b, err := GrapesJSComponentModelJSON(v)
	if err != nil {
		panic(err)
	}
	return b
}

// GrapesJSComponentViewJSON marshals a view definition object.
func GrapesJSComponentViewJSON(v any) (json.RawMessage, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	return b, nil
}

// GrapesJSComponentMustViewJSON marshals a view definition or panics.
func GrapesJSComponentMustViewJSON(v any) json.RawMessage {
	b, err := GrapesJSComponentViewJSON(v)
	if err != nil {
		panic(err)
	}
	return b
}
