package components

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"strings"

	"github.com/lariv-in/lariv/getters"
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
	// Attr is an optional Getter returning additional HTML attributes to apply to the input.
	Attr getters.Getter[HTMLAttributes]
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
func (e InputNullableText) Build(cat Catalog, ctx context.Context, w io.Writer) error {
	value := ""
	if e.Getter != nil {
		v, err := e.Getter(ctx)
		if err != nil {
			slog.Error("InputNullableText getter failed", "error", err, "key", e.Key)
		} else if v != nil {
			value = *v
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
		slog.Error("InputNullableText attr getter failed", "error", err, "key", e.Key)
		attrs = HTMLAttributes{}
	}
	return Execute(w, "input_nullable_text", struct {
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
