package components

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"strings"
	"time"

	"github.com/lariv-in/lariv/getters"
)

// InputDuration represents a text input field component designed to capture time duration strings.
// It accepts valid Go time duration strings (e.g., "30s", "15m", "2h45m") and parses them into a *time.Duration pointer.
//
// Use Cases:
//   - Defining timeout configurations (e.g., connection keep-alive timeouts, lockout delay policies).
//   - Scheduling time windows or interval execution policies.
//
// Example:
//
//	&components.InputDuration{
//	    Label:  "Lockout Duration",
//	    Name:   "lockout_dur",
//	    Getter: getters.Key[*time.Duration]("$in.LockoutDuration"),
//	}
type InputDuration struct {
	// Page embeds common component properties like Key and Roles.
	Page
	// Label represents the header label text displayed above the text input.
	Label string
	// Name represents the HTML form parameter name attribute.
	Name string
	// Getter is the dynamic function retrieving the default/current *time.Duration pointer value.
	Getter getters.Getter[*time.Duration]
	// Required is a boolean indicating if this form duration is a mandatory input.
	Required bool
	// Classes represents additional CSS classes applied to the output HTML wrapper.
	// (Discouraged: Use layout containers or theme styling instead of custom styling overrides).
	Classes string
	// Hidden specifies if this duration field is rendered as a hidden form element instead of an interactive text box.
	Hidden bool
	// Attr is an optional Getter returning additional HTML attributes to apply to the input.
	Attr getters.Getter[HTMLAttributes]
}

// GetKey returns the unique key identifier for this InputDuration component.
func (e InputDuration) GetKey() string {
	return e.Key
}

// GetRoles returns the authorized roles required to view this InputDuration.
func (e InputDuration) GetRoles() []string {
	return e.Roles
}

// Build compiles the InputDuration component into a Div wrapping a text Input field.
func (e InputDuration) Build(cat Catalog, ctx context.Context, w io.Writer) error {
	value := ""
	if e.Getter != nil {
		d, err := e.Getter(ctx)
		if err != nil {
			slog.Error("InputDuration getter failed", "error", err, "key", e.Key)
		} else if d != nil {
			value = d.String()
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
		slog.Error("InputDuration attr getter failed", "error", err, "key", e.Key)
		attrs = HTMLAttributes{}
	}
	return Execute(w, "input_duration", struct {
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

// Parse extracts and parses standard duration strings into a pointer to a time.Duration.
func (e InputDuration) Parse(v any, _ context.Context) (any, error) {
	vals, _ := v.([]string)
	if len(vals) == 0 {
		return (*time.Duration)(nil), nil
	}
	raw := strings.TrimSpace(vals[0])
	if raw == "" {
		return (*time.Duration)(nil), nil
	}
	d, err := time.ParseDuration(raw)
	if err != nil {
		return nil, fmt.Errorf("invalid duration")
	}
	return &d, nil
}

// GetName returns the HTML form element's name attribute value.
func (e InputDuration) GetName() string {
	return e.Name
}
