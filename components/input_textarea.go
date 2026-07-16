package components

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/lariv-in/lariv/getters"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

// InputTextarea represents a multi-line text input form field component.
// It renders an HTML textarea element with configurable rows height (defaults to 3 rows).
//
// Use Cases:
//   - Capturing longer, multi-line user text entries such as biographies, descriptions, comments, or street addresses.
//
// Example:
//
//	&components.InputTextarea{
//	    Label:  "Item Description",
//	    Name:   "description",
//	    Rows:   5,
//	}
type InputTextarea struct {
	// Page embeds common component properties like Key and Roles.
	Page
	// Label represents the header label text displayed above the textarea input.
	Label string
	// Name represents the HTML form parameter name attribute.
	Name string
	// Getter is the dynamic function retrieving the default/current textarea text value.
	Getter getters.Getter[string]
	// Required is a boolean indicating if this textarea is a mandatory input.
	Required bool
	// Rows specifies the vertical height of the textarea element in text rows (defaults to 3).
	Rows int
	// Classes represents additional CSS classes applied to the output HTML wrapper.
	// (Discouraged: Use layout containers or theme styling instead of custom styling overrides).
	Classes string
}

// GetKey returns the unique key identifier for this InputTextarea component.
func (e InputTextarea) GetKey() string {
	return e.Key
}

// GetRoles returns the authorized roles required to view this InputTextarea.
func (e InputTextarea) GetRoles() []string {
	return e.Roles
}

// Build compiles the InputTextarea component into a Div wrapping an HTML Textarea.
func (e InputTextarea) Build(ctx context.Context) Node {
	rows := e.Rows
	if rows <= 0 {
		rows = 3
	}
	var valueNode Node = Text("")
	if e.Getter != nil {
		value, err := e.Getter(ctx)
		if err != nil {
			slog.Error("InputTextarea getter failed", "error", err, "key", e.Key)
		} else {
			valueNode = Text(value)
		}
	}
	return Div(
		Class(fmt.Sprintf("my-1 %s", e.Classes)),
		Label(
			Class("label text-sm font-bold flex flex-col items-start gap-1"),
			Text(e.Label),
			Textarea(Name(e.Name),
				Rows(fmt.Sprintf("%d", rows)),
				valueNode,
				Class(fmt.Sprintf("textarea textarea-bordered w-full %s", e.Classes)),
				If(e.Required, Required())),
		),
	)
}

// Parse extracts the textarea string value from parameters.
func (e InputTextarea) Parse(v any, _ context.Context) (any, error) {
	vals, _ := v.([]string)
	if len(vals) == 0 {
		return "", nil
	}
	return vals[0], nil
}

// GetName returns the HTML form element's name attribute value.
func (e InputTextarea) GetName() string {
	return e.Name
}
