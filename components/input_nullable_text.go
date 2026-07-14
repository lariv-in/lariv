package components

import (
	"context"
	"fmt"
	"log/slog"
	"reflect"
	"strings"

	"github.com/lariv-in/lago/getters"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

// InputNullableText represents a text input form field component designed to bind to optional pointer strings (*string).
// If empty value is submitted, Parse returns a nil string pointer ensuring GORM updates the database field to SQL NULL instead of empty text.
//
// Use Cases:
//   - Handling optional model text properties (e.g., middle names, secondary address lines, or non-mandatory description fields).
//
// Example:
//
//	&components.InputNullableText{
//	    Label:  "Middle Name",
//	    Name:   "middle_name",
//	    Getter: getters.Key[*string]("$in.MiddleName"),
//	}
type InputNullableText struct {
	// Page embeds common component properties like Key and Roles.
	Page
	// Label represents the header label text displayed above the text input.
	Label string
	// Name represents the HTML form parameter name attribute.
	Name string
	// Getter is the dynamic function retrieving the default/current *string pointer value.
	Getter getters.Getter[*string]
	// Required specifies if inputting text is mandatory.
	Required bool
	// Classes represents additional CSS classes applied to the output HTML wrapper.
	// (Discouraged: Use layout containers or theme styling instead of custom styling overrides).
	Classes string
	// Hidden specifies if this text field is rendered as a hidden input element.
	Hidden bool
	// Attr is an optional Getter returning additional HTML nodes/attributes to apply to the input.
	Attr getters.Getter[Node]
}

// GetKey returns the unique key identifier for this InputNullableText component.
func (e InputNullableText) GetKey() string {
	return e.Key
}

// GetRoles returns the authorized roles required to view this InputNullableText.
func (e InputNullableText) GetRoles() []string {
	return e.Roles
}

// Build compiles the InputNullableText component into a Div wrapping a text input.
func (e InputNullableText) Build(ctx context.Context) Node {
	var valueNode Node = Value("")
	if e.Getter != nil {
		value, err := e.Getter(ctx)
		if err != nil {
			slog.Error("InputNullableText getter failed", "error", err, "key", e.Key)
		} else if value != nil {
			valueNode = Value(*value)
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
							slog.Error("InputNullableText attr getter panicked", "panic", r, "key", e.Key)
						}
					}()
					n, err := e.Attr(ctx)
					if err != nil {
						slog.Error("InputNullableText attr getter failed", "error", err, "key", e.Key)
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

// Parse extracts text strings from parameters and returns a string pointer or a nil pointer if empty.
func (e InputNullableText) Parse(v any, _ context.Context) (any, error) {
	vals, _ := v.([]string)
	if len(vals) == 0 {
		return (*string)(nil), nil
	}
	raw := strings.TrimSpace(vals[0])
	if raw == "" {
		return (*string)(nil), nil
	}
	return &raw, nil
}

// GetName returns the HTML form element's name attribute value.
func (e InputNullableText) GetName() string {
	return e.Name
}
