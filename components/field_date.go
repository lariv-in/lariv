package components

import (
	"context"
	"log/slog"
	"time"

	"github.com/lariv-in/lago/getters"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

// FieldDate represents a read-only date display field.
// It formats and outputs a time.Time value in the user's localized timezone (retrieved from the "$tz" context key, falling back to DefaultTimeZone).
// It renders an HTML <time> tag with the date string (YYYY-MM-DD format).
//
// Use Cases:
//   - Displaying localized dates in invoice headers, audit logs, or profile summaries.
//
// Example:
//
//	&components.FieldDate{
//	    Getter: getters.Key[time.Time]("$in.CreatedAt"),
//	}
type FieldDate struct {
	// Page embeds common component properties like Key and Roles.
	Page
	// Getter is the dynamic function retrieving the time.Time value to display.
	Getter getters.Getter[time.Time]
	// Classes represents additional CSS classes applied to the output HTML time element.
	// (Discouraged: Use layout containers or theme styling instead of custom styling overrides).
	Classes string
}

// GetKey returns the unique key identifier for this FieldDate component.
func (e FieldDate) GetKey() string {
	return e.Key
}

// GetRoles returns the authorized roles required to view this FieldDate.
func (e FieldDate) GetRoles() []string {
	return e.Roles
}

// Build compiles the FieldDate component into a Time HTML Node formatted according to the context timezone.
func (e FieldDate) Build(ctx context.Context) Node {
	if e.Getter == nil {
		return Group{}
	}
	v, err := e.Getter(ctx)
	if err != nil {
		slog.Error("FieldDate getter failed", "error", err, "key", e.Key)
		return ContainerError{Error: getters.Static(err)}.Build(ctx)
	}
	timezone, _ := ctx.Value("$tz").(*time.Location)
	if timezone == nil {
		timezone = DefaultTimeZone
	}
	return Time(Class(e.Classes), DateTime(v.In(timezone).Format(time.DateOnly)), Text(v.In(timezone).Format(time.DateOnly)))
}
