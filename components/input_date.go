package components

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/lariv-in/lariv/getters"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

// InputDate represents a date selection input form field component.
// It renders an HTML date picker input and formats default values into "YYYY-MM-DD" matching the context's localized timezone ($tz).
//
// Use Cases:
//   - Selecting specific dates for form inputs (e.g., employee birthdate, project target deadlines, contract starting dates).
//
// Example:
//
//	&components.InputDate{
//	    Label:  "Target Launch Date",
//	    Name:   "target_launch",
//	    Getter: getters.Key[time.Time]("$in.TargetLaunch"),
//	}
type InputDate struct {
	// Page embeds common component properties like Key and Roles.
	Page
	// Label represents the header label text displayed above the date picker input.
	Label string
	// Name represents the HTML form parameter name attribute.
	Name string
	// Getter is the dynamic function retrieving the default/current time.Time value.
	Getter getters.Getter[time.Time]
	// Required is a boolean indicating if this form date is a mandatory input.
	Required bool
	// Classes represents additional CSS classes applied to the output HTML wrapper.
	// (Discouraged: Use layout containers or theme styling instead of custom styling overrides).
	Classes string
	// Hidden specifies if this date field is rendered as a hidden form element instead of an interactive calendar picker.
	Hidden bool
}

// GetKey returns the unique key identifier for this InputDate component.
func (e InputDate) GetKey() string {
	return e.Key
}

// GetRoles returns the authorized roles required to view this InputDate.
func (e InputDate) GetRoles() []string {
	return e.Roles
}

// Build compiles the InputDate component into a Div wrapping a date selection Input.
func (e InputDate) Build(ctx context.Context) Node {
	timezone, _ := ctx.Value("$tz").(*time.Location)
	if timezone == nil {
		timezone = DefaultTimeZone
	}
	var valueNode Node = Value("")
	if e.Getter != nil {
		t, err := e.Getter(ctx)
		if err != nil {
			slog.Error("InputDate getter failed", "error", err, "key", e.Key)
		} else if !t.IsZero() {
			valueNode = Value(t.In(timezone).Format("2006-01-02"))
		}
	}
	if e.Hidden {
		return Div(
			Class("hidden"),
			Input(Type("hidden"), Name(e.Name), valueNode),
		)
	}
	return Div(
		Class(fmt.Sprintf("my-1 %s", e.Classes)),
		Label(
			Class("label text-sm font-bold flex flex-col items-start gap-1"),
			Text(e.Label),
			Input(Type("date"), Name(e.Name), valueNode, Class(fmt.Sprintf("input input-bordered w-full %s", e.Classes)), If(e.Required, Required())),
		),
	)
}

// Parse extracts the selected date string from parameter arrays and parses it as a time.Time in localized location.
func (e InputDate) Parse(v any, ctx context.Context) (any, error) {
	vals, _ := v.([]string)
	if len(vals) == 0 || vals[0] == "" {
		return nil, nil
	}
	timezone, _ := ctx.Value("$tz").(*time.Location)
	if timezone == nil {
		timezone = DefaultTimeZone
	}
	return time.ParseInLocation("2006-01-02", vals[0], timezone)
}

// GetName returns the HTML form element's name attribute value.
func (e InputDate) GetName() string {
	return e.Name
}
