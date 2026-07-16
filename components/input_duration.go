package components

import (
	"context"
	"fmt"
	"log/slog"
	"reflect"
	"strings"
	"time"

	"github.com/lariv-in/lariv/getters"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
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
	// Attr is an optional Getter returning additional HTML nodes/attributes to apply to the input.
	Attr getters.Getter[Node]
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
func (e InputDuration) Build(ctx context.Context) Node {
	var valueNode Node = Value("")
	if e.Getter != nil {
		d, err := e.Getter(ctx)
		if err != nil {
			slog.Error("InputDuration getter failed", "error", err, "key", e.Key)
		} else if d != nil {
			valueNode = Value(d.String())
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
							slog.Error("InputDuration attr getter panicked", "panic", r, "key", e.Key)
						}
					}()
					n, err := e.Attr(ctx)
					if err != nil {
						slog.Error("InputDuration attr getter failed", "error", err, "key", e.Key)
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
