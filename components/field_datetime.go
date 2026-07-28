package components

import (
	"context"
	"io"
	"log/slog"
	"time"

	"github.com/lariv-in/lariv/getters"
)

// FieldDatetime represents a read-only date and time display field.
// It formats and outputs a time.Time value including timezone-localized hours, minutes, and seconds in the "Mon, 02 Jan 2006 15:04:05" format.
//
// Use Cases:
//   - Showing full creation, modification, or session scheduling timestamps.
//
// Example:
//
//	&components.FieldDatetime{
//	    Getter: getters.Key[time.Time]("$in.UpdatedAt"),
//	}
type FieldDatetime struct {
	// Page embeds common component properties like Key and Roles.
	Page
	// Getter is the dynamic function retrieving the time.Time value to display.
	Getter getters.Getter[time.Time]
	// Classes represents additional CSS classes applied to the output HTML div wrapper.
	// (Discouraged: Use layout containers or theme styling instead of custom styling overrides).
	Classes string
}

// GetKey returns the unique key identifier for this FieldDatetime component.
func (e FieldDatetime) GetKey() string {
	return e.Key
}

// GetRoles returns the authorized roles required to view this FieldDatetime.
func (e FieldDatetime) GetRoles() []string {
	return e.Roles
}

// Build compiles the FieldDatetime component into a Div containing the localized formatted time string.
func (e FieldDatetime) Build(cat Catalog, ctx context.Context, w io.Writer) error {
	if e.Getter == nil {
		return nil
	}
	v, err := e.Getter(ctx)
	if err != nil {
		slog.Error("FieldDatetime getter failed", "error", err, "key", e.Key)
		return ContainerError{Error: getters.Static(err)}.Build(cat, ctx, w)
	}
	timezone, _ := ctx.Value("$tz").(*time.Location)
	if timezone == nil {
		timezone = DefaultTimeZone
	}
	return Execute(w, "field_datetime", struct {
		Classes string
		Value   string
	}{Classes: e.Classes, Value: v.In(timezone).Format("Mon, 02 Jan 2006 15:04:05")})
}
