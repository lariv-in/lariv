package components

import (
	"context"
	"io"
	"log/slog"
	"time"

	"github.com/lariv-in/lariv/getters"
)

// InputDatetime represents a date and time selection input form field component.
// It renders an HTML datetime-local picker input and formats default values into "YYYY-MM-DDTHH:MM" matching the context's localized timezone ($tz).
//
// Use Cases:
//   - Booking calendar appointments, setting precise project expiration hours, or scheduling event start/end timestamps.
//
// Example:
//
//	&components.InputDatetime{
//	    Label:  "Event Time",
//	    Name:   "event_time",
//	    Getter: getters.Key[time.Time]("$in.EventTime"),
//	}
type InputDatetime struct {
	// Page embeds common component properties like Key and Roles.
	Page
	// Label represents the header label text displayed above the datetime picker input.
	Label string
	// Name represents the HTML form parameter name attribute.
	Name string
	// Getter is the dynamic function retrieving the default/current time.Time value.
	Getter getters.Getter[time.Time]
	// Required is a boolean indicating if this form datetime is a mandatory input.
	Required bool
	// Classes represents additional CSS classes applied to the output HTML wrapper.
	// (Discouraged: Use layout containers or theme styling instead of custom styling overrides).
	Classes string
	// Hidden specifies if this datetime field is rendered as a hidden form element instead of an interactive picker.
	Hidden bool
}

// GetKey returns the unique key identifier for this InputDatetime component.
func (e InputDatetime) GetKey() string {
	return e.Key
}

// GetRoles returns the authorized roles required to view this InputDatetime.
func (e InputDatetime) GetRoles() []string {
	return e.Roles
}

// Build compiles the InputDatetime component into a Div wrapping a datetime-local input element.
func (e InputDatetime) Build(cat Catalog, ctx context.Context, w io.Writer) error {
	timezone, _ := ctx.Value("$tz").(*time.Location)
	if timezone == nil {
		timezone = DefaultTimeZone
	}
	value := ""
	if e.Getter != nil {
		t, err := e.Getter(ctx)
		if err != nil {
			slog.Error("InputDatetime getter failed", "error", err, "key", e.Key)
		} else if !t.IsZero() {
			value = t.In(timezone).Format("2006-01-02T15:04")
		}
	}
	return Execute(w, "input_datetime", struct {
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

// Parse extracts the selected datetime string from parameter arrays and parses it as a time.Time in localized location.
func (e InputDatetime) Parse(v any, ctx context.Context) (any, error) {
	vals, _ := v.([]string)
	if len(vals) == 0 || vals[0] == "" {
		return nil, nil
	}
	timezone, _ := ctx.Value("$tz").(*time.Location)
	if timezone == nil {
		timezone = DefaultTimeZone
	}
	return time.ParseInLocation("2006-01-02T15:04", vals[0], timezone)
}

// GetName returns the HTML form element's name attribute value.
func (e InputDatetime) GetName() string {
	return e.Name
}
