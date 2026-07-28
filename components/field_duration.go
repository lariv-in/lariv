package components

import (
	"context"
	"io"
	"log/slog"
	"time"

	"github.com/lariv-in/lariv/getters"
)

// FieldDuration represents a read-only field that displays time duration values.
// Resolves a *time.Duration pointer from the context and outputs its standard string format (e.g. "1h30m0s").
//
// Use Cases:
//   - Showing time intervals or processing runtimes (e.g. "Elapsed Time", "Job Duration", "Lockout Timeout").
//
// Example:
//
//	&components.FieldDuration{
//	    Getter: getters.Key[*time.Duration]("$in.TimeoutInterval"),
//	}
type FieldDuration struct {
	// Page embeds common component properties like Key and Roles.
	Page
	// Getter is the dynamic function retrieving the *time.Duration pointer to display.
	Getter getters.Getter[*time.Duration]
	// Classes represents additional CSS classes applied to the output HTML div wrapper.
	// (Discouraged: Use layout containers or theme styling instead of custom styling overrides).
	Classes string
}

// GetKey returns the unique key identifier for this FieldDuration component.
func (e FieldDuration) GetKey() string {
	return e.Key
}

// GetRoles returns the authorized roles required to view this FieldDuration.
func (e FieldDuration) GetRoles() []string {
	return e.Roles
}

// Build compiles the FieldDuration component into a Div displaying the duration string.
// Includes a defer/recover safety net to handle getter panics cleanly by writing an empty string.
func (e FieldDuration) Build(cat Catalog, ctx context.Context, w io.Writer) (err error) {
	if e.Getter == nil {
		return nil
	}
	defer func() {
		if r := recover(); r != nil {
			slog.Error("FieldDuration getter panicked", "panic", r, "key", e.Key)
			err = Execute(w, "field_duration", struct {
				Classes string
				Value   string
			}{Classes: e.Classes, Value: ""})
		}
	}()
	v, err := e.Getter(ctx)
	if err != nil {
		slog.Error("FieldDuration getter failed", "error", err, "key", e.Key)
		return ContainerError{Error: getters.Static(err)}.Build(cat, ctx, w)
	}
	value := ""
	if v != nil {
		value = v.String()
	}
	return Execute(w, "field_duration", struct {
		Classes string
		Value   string
	}{Classes: e.Classes, Value: value})
}
