package components

import (
	"context"
	"fmt"
	"log/slog"
	"reflect"

	"github.com/lariv-in/lariv/getters"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
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
	// Attr is an optional Getter returning additional HTML nodes/attributes to apply to the input.
	Attr getters.Getter[Node]
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
func (e InputText) Build(ctx context.Context) Node {
	valueNode := Value("")
	if e.Getter != nil {
		value, err := e.Getter(ctx)
		if err != nil {
			slog.Error("InputText getter failed", "error", err, "key", e.Key)
		} else {
			valueNode = Value(value)
		}
	}

	wrapClass := fmt.Sprintf("my-1 %s", e.Classes)
	if e.Hidden {
		wrapClass += " hidden"
	}
	return Div(
		Class(wrapClass),
		Label(
			Class("label text-sm font-bold flex flex-col items-start gap-1"),
			If(!e.Hidden, Text(e.Label)),
			Input(
				If(!e.Hidden, Type("text")), If(e.Hidden, Type("hidden")), Name(e.Name),
				valueNode,
				Class(fmt.Sprintf("input input-bordered w-full %s", e.Classes)),
				If(e.Required, Required()),
				Iff(e.Attr != nil, func() (out Node) {
					out = Raw("")
					defer func() {
						if r := recover(); r != nil {
							slog.Error("InputText attr getter panicked", "panic", r, "key", e.Key)
						}
					}()
					n, err := e.Attr(ctx)
					if err != nil {
						slog.Error("InputText attr getter failed", "error", err, "key", e.Key)
						return out
					}
					if n == nil {
						return out
					}
					v := reflect.ValueOf(n)
					if (v.Kind() == reflect.Pointer || v.Kind() == reflect.Map || v.Kind() == reflect.Slice || v.Kind() == reflect.Interface || v.Kind() == reflect.Func) && v.IsNil() {
						return out
					}
					return n
				}),
			),
		),
	)
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
