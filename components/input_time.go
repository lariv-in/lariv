package components

import (
	"context"
	"io"
	"log/slog"
	"time"

	"github.com/lariv-in/lariv/getters"
)

// InputTime represents a clock time input form field component.
// It renders an HTML time input (`<input type="time">`) formatted as "15:04" (24-hour time), adjusting for timezone contexts.
//
// Use Cases:
//   - Inputting daily schedules, business opening/closing hours, or clock alarms (e.g. shift start time).
//
// Example:
//
//	&components.InputTime{
//	    Label:  "Daily Meeting Time",
//	    Name:   "meeting_time",
//	    Getter: getters.Key[time.Time]("$in.MeetingTime"),
//	}
type InputTime struct {
	// Page embeds common component properties like Key and Roles.
	Page
	// Label represents the header label text displayed above the time input.
	Label string
	// Name represents the HTML form parameter name attribute.
	Name string
	// Getter is the dynamic function retrieving the default/current time.Time clock value.
	Getter getters.Getter[time.Time]
	// Required is a boolean indicating if this form time is a mandatory input.
	Required bool
	// Classes represents additional CSS classes applied to the output HTML wrapper.
	// (Discouraged: Use layout containers or theme styling instead of custom styling overrides).
	Classes string
	// Hidden specifies if this time field is rendered as a hidden input element.
	Hidden bool
}

// Build compiles the InputTime component into a Div wrapping a clock time input Node, formatting with target timezones.
func (e InputTime) Build(cat Catalog, ctx context.Context, w io.Writer) error {
	timezone, _ := ctx.Value("$tz").(*time.Location)
	if timezone == nil {
		timezone = DefaultTimeZone
	}
	value := ""
	if e.Getter != nil {
		t, err := e.Getter(ctx)
		if err != nil {
			slog.Error("InputTime getter failed", "error", err, "key", e.Key)
		} else if !t.IsZero() {
			value = t.In(timezone).Format("15:04")
		}
	}
	return Execute(w, "input_time", struct {
		Hidden   bool
		Classes  string
		Label    string
		Name     string
		Value    string
		Required bool
	}{
		Hidden:   e.Hidden,
		Classes:  e.Classes,
		Label:    e.Label,
		Name:     e.Name,
		Value:    value,
		Required: e.Required,
	})
}

// Parse extracts text values and parses them in the context's target location timezone.
func (e InputTime) Parse(v any, ctx context.Context) (any, error) {
	vals, _ := v.([]string)
	if len(vals) == 0 || vals[0] == "" {
		return nil, nil
	}
	timezone, _ := ctx.Value("$tz").(*time.Location)
	if timezone == nil {
		timezone = DefaultTimeZone
	}
	return time.ParseInLocation("15:04", vals[0], timezone)
}

// GetKey returns the unique key identifier for this InputTime component.
func (e InputTime) GetKey() string {
	return e.Key
}

// GetRoles returns the authorized roles required to view this InputTime.
func (e InputTime) GetRoles() []string {
	return e.Roles
}

// GetName returns the HTML form element's name attribute value.
func (e InputTime) GetName() string {
	return e.Name
}
