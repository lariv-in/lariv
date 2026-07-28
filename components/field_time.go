package components

import (
	"context"
	"io"
	"log/slog"
	"time"

	"github.com/lariv-in/lariv/getters"
)

// FieldTime represents a read-only localized time display field.
// It formats and outputs only the time portion (HH:MM:SS format using time.TimeOnly) of a time.Time value,
// localized to the context's timezone ($tz).
//
// Use Cases:
//   - Showing specific time-of-day information, such as store opening hours, appointment starting times, or precise logs.
//
// Example:
//
//	&components.FieldTime{
//	    Getter: getters.Key[time.Time]("$in.OpenTime"),
//	}
type FieldTime struct {
	// Page embeds common component properties like Key and Roles.
	Page
	// Getter is the dynamic function retrieving the time.Time value to display.
	Getter getters.Getter[time.Time]
	// Classes represents additional CSS classes applied to the output HTML div wrapper.
	// (Discouraged: Use layout containers or theme styling instead of custom styling overrides).
	Classes string
}

// GetKey returns the unique key identifier for this FieldTime component.
func (e FieldTime) GetKey() string {
	return e.Key
}

// GetRoles returns the authorized roles required to view this FieldTime.
func (e FieldTime) GetRoles() []string {
	return e.Roles
}

// Build compiles the FieldTime component into a Div containing the localized formatted time string.
func (e FieldTime) Build(cat Catalog, ctx context.Context, w io.Writer) error {
	if e.Getter == nil {
		return nil
	}
	v, err := e.Getter(ctx)
	if err != nil {
		slog.Error("FieldTime getter failed", "error", err, "key", e.Key)
		return ContainerError{Error: getters.Static(err)}.Build(cat, ctx, w)
	}
	timezone, _ := ctx.Value("$tz").(*time.Location)
	if timezone == nil {
		timezone = DefaultTimeZone
	}
	return Execute(w, "field_time", struct {
		Classes string
		Value   string
	}{Classes: e.Classes, Value: v.In(timezone).Format(time.TimeOnly)})
}
