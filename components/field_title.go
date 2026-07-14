package components

import (
	"context"
	"log/slog"

	"github.com/lariv-in/lago/getters"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

// FieldTitle represents a read-only title header display field.
// It resolves a string value dynamically from context and renders it as primary-colored title text ("text-xl font-semibold text-primary").
//
// Use Cases:
//   - Showing main page section titles, panel headers, or modal block titles.
//
// Example:
//
//	&components.FieldTitle{
//	    Getter: getters.Static("Billing Overview"),
//	}
type FieldTitle struct {
	// Page embeds common component properties like Key and Roles.
	Page
	// Getter is the dynamic function retrieving the title string.
	Getter getters.Getter[string]
	// Classes represents additional CSS classes applied to the output HTML div wrapper.
	// (Discouraged: Use layout containers or theme styling instead of custom styling overrides).
	Classes string
}

// GetKey returns the unique key identifier for this FieldTitle component.
func (e FieldTitle) GetKey() string {
	return e.Key
}

// GetRoles returns the authorized roles required to view this FieldTitle.
func (e FieldTitle) GetRoles() []string {
	return e.Roles
}

// Build compiles the FieldTitle component into a Div Node styled as primary header text.
func (e FieldTitle) Build(ctx context.Context) Node {
	value := ""
	if e.Getter != nil {
		v, err := e.Getter(ctx)
		if err != nil {
			slog.Error("FieldTitle getter failed", "error", err, "key", e.Key)
			return ContainerError{Error: getters.Static(err)}.Build(ctx)
		}
		value = v
	}
	return Div(Class("text-xl font-semibold text-primary "+e.Classes), Text(value))
}
