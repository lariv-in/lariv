package components

import (
	"context"
	"fmt"
	"io"
	"log/slog"

	"github.com/lariv-in/lariv/getters"
)

// InputText represents a standard text input form field component.
// It renders an HTML text input (`<input type="text">`) or a hidden input depending on Hidden configuration.
//
// Use Cases:
//   - Capturing short text entries like first names, last names, usernames, job titles, or text identifiers.
//
// Example:
//
//	&components.InputText{
//	    Label:    "Display Name",
//	    Name:     "display_name",
//	    Required: true,
//	}
type InputText struct {
	// Page embeds common component properties like Key and Roles.
	Page
	// Label represents the header label text displayed above the text input.
	Label string
	// Name represents the HTML form parameter name attribute.
	Name string
	// Getter is the dynamic function retrieving the default/current string value.
	Getter getters.Getter[string]
	// Attr is an optional Getter returning additional HTML attributes to apply to the input.
	Attr getters.Getter[HTMLAttributes]
	// Required is a boolean indicating if this form text is a mandatory input.
	Required bool
	// Classes represents additional CSS classes applied to the output HTML wrapper.
	// (Discouraged: Use layout containers or theme styling instead of custom styling overrides).
	Classes string
	// Hidden specifies if this text field is rendered as a hidden input element.
	Hidden bool
}

// GetKey returns the unique key identifier for this InputText component.
func (e InputText) GetKey() string {
	return e.Key
}

// GetRoles returns the authorized roles required to view this InputText.
func (e InputText) GetRoles() []string {
	return e.Roles
}

// Build compiles the InputText component into a Div wrapping a text/hidden Input.
func (e InputText) Build(cat Catalog, ctx context.Context, w io.Writer) error {
	value := ""
	if e.Getter != nil {
		v, err := e.Getter(ctx)
		if err != nil {
			slog.Error("InputText getter failed", "error", err, "key", e.Key)
		} else {
			value = v
		}
	}

	wrapClass := fmt.Sprintf("my-1 %s", e.Classes)
	if e.Hidden {
		wrapClass += " hidden"
	}
	inputType := "text"
	if e.Hidden {
		inputType = "hidden"
	}
	attrs, err := ResolveAttrs(ctx, e.Attr)
	if err != nil {
		slog.Error("InputText attr getter failed", "error", err, "key", e.Key)
		attrs = HTMLAttributes{}
	}
	return Execute(w, "input_text", struct {
		WrapClass string
		Hidden    bool
		Label     string
		Type      string
		Name      string
		Value     string
		Classes   string
		Required  bool
		Attrs     HTMLAttributes
	}{
		WrapClass: wrapClass,
		Hidden:    e.Hidden,
		Label:     e.Label,
		Type:      inputType,
		Name:      e.Name,
		Value:     value,
		Classes:   e.Classes,
		Required:  e.Required,
		Attrs:     attrs,
	})
}

// Parse extracts the text string value from input parameters.
func (e InputText) Parse(v any, _ context.Context) (any, error) {
	vals, _ := v.([]string)
	if len(vals) == 0 {
		return "", nil
	}
	return vals[0], nil
}

// GetName returns the HTML form element's name attribute value.
func (e InputText) GetName() string {
	return e.Name
}
