package components

import (
	"context"
	"io"
	"log/slog"
	"strconv"

	"github.com/lariv-in/lariv/getters"
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
	// Attr is an optional Getter returning additional HTML attributes to apply to the input.
	Attr getters.Getter[HTMLAttributes]
}

// Build compiles the InputCheckbox component into a wrapper Div Node with nested checkbox Input.
func (e InputCheckbox) Build(cat Catalog, ctx context.Context, w io.Writer) error {
	checked := false
	if e.Getter != nil {
		value, err := e.Getter(ctx)
		if err != nil {
			slog.Error("InputCheckbox getter failed", "error", err, "key", e.Key)
		} else {
			checked = value
		}
	}
	if e.Hidden {
		return Execute(w, "input_checkbox", struct {
			Hidden  bool
			Name    string
			Value   string
			Classes string
			Label   string
			XModel  string
			Checked bool
			Attrs   HTMLAttributes
		}{
			Hidden: true,
			Name:   e.Name,
			Value:  strconv.FormatBool(checked),
		})
	}
	attrs, err := ResolveAttrs(ctx, e.Attr)
	if err != nil {
		slog.Error("InputCheckbox Attr getter failed", "error", err, "key", e.Key)
		attrs = HTMLAttributes{}
	}
	return Execute(w, "input_checkbox", struct {
		Hidden  bool
		Name    string
		Value   string
		Classes string
		Label   string
		XModel  string
		Checked bool
		Attrs   HTMLAttributes
	}{
		Hidden:  false,
		Name:    e.Name,
		Classes: e.Classes,
		Label:   e.Label,
		XModel:  e.XModel,
		Checked: checked,
		Attrs:   attrs,
	})
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
