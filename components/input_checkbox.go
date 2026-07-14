package components

import (
	"context"
	"log/slog"
	"strconv"

	"github.com/lariv-in/lago/getters"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

// InputCheckbox represents a boolean state toggler form input component.
// It renders an HTML checkbox input alongside its label string, and integrates with Alpine.js data models if configured.
//
// Use Cases:
//   - Toggling binary preferences (e.g., agreeing to Terms of Service, enabling push notifications, opting in to newsletters).
//
// Example:
//
//	&components.InputCheckbox{
//	    Label:  "Accept Marketing Emails",
//	    Name:   "accept_marketing",
//	    Getter: getters.Key[bool]("$in.AcceptMarketing"),
//	}
type InputCheckbox struct {
	// Page embeds common component properties like Key and Roles.
	Page
	// Label represents the text label string displayed next to the checkbox.
	Label string
	// Name represents the HTML form parameter name attribute.
	Name string
	// Getter is the dynamic function retrieving the default/current checked state.
	Getter getters.Getter[bool]
	// XModel is an optional string specifying an Alpine.js x-model attribute binding.
	XModel string
	// Required is a boolean indicating if this form checkbox is a mandatory input.
	Required bool
	// Classes represents additional CSS classes applied to the output HTML wrapper.
	// (Discouraged: Use layout containers or theme styling instead of custom styling overrides).
	Classes string
	// Hidden specifies if this checkbox is rendered as a hidden form element instead of an interactive toggle.
	Hidden bool
	// Attr is an optional Getter returning additional HTML nodes/attributes to apply to the input.
	Attr getters.Getter[Node]
}

// Build compiles the InputCheckbox component into a wrapper Div Node with nested checkbox Input.
func (e InputCheckbox) Build(ctx context.Context) Node {
	checked := false
	var checkedNode Node = Raw("")
	if e.Getter != nil {
		value, err := e.Getter(ctx)
		if err != nil {
			slog.Error("InputCheckbox getter failed", "error", err, "key", e.Key)
		} else {
			checked = value
			if checked {
				checkedNode = Checked()
			}
		}
	}
	if e.Hidden {
		return Div(
			Class("hidden"),
			Input(
				Type("hidden"),
				Name(e.Name),
				Value(strconv.FormatBool(checked)),
			),
		)
	}
	return Div(
		Class(e.Classes),
		Label(
			Class("label text-sm font-bold cursor-pointer justify-start gap-2 flex flex-row items-center"),
			Input(
				Type("checkbox"),
				If(e.Name != "", Name(e.Name)),
				Value("true"),
				Class("checkbox"),
				If(e.XModel != "", Attr("x-model", e.XModel)),
				checkedNode,
				Iff(e.Attr != nil, func() Node {
					n, err := e.Attr(ctx)
					if err != nil {
						slog.Error("InputCheckbox Attr getter failed", "error", err, "key", e.Key)
						return Raw("")
					}
					return n
				}),
			),
			Span(Class("label-text"), Text(e.Label)),
		),
	)
}

// Parse extracts and parses the boolean checked status from request parameter strings.
func (e InputCheckbox) Parse(v any, _ context.Context) (any, error) {
	vals, _ := v.([]string)
	if len(vals) == 0 {
		return false, nil
	}
	return strconv.ParseBool(vals[0])
}

// GetKey returns the unique key identifier for this InputCheckbox component.
func (e InputCheckbox) GetKey() string {
	return e.Key
}

// GetRoles returns the authorized roles required to view this InputCheckbox.
func (e InputCheckbox) GetRoles() []string {
	return e.Roles
}

// GetName returns the HTML form element's name attribute value.
func (e InputCheckbox) GetName() string {
	return e.Name
}
