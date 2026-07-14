package components

import (
	"context"
	"log/slog"

	"github.com/lariv-in/lago/getters"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

// FieldText represents a read-only plain-text display field.
// It resolves a string value dynamically from context and renders it inside a wrapping HTML div element.
//
// Use Cases:
//   - Showing basic textual model properties in details pages or summary sections (e.g., first/last names, company names, item categories).
//
// Example:
//
//	&components.FieldText{
//	    Getter:  getters.Key[string]("$in.CustomerName"),
//	    Classes: "font-medium text-base-content",
//	}
type FieldText struct {
	// Page embeds common component properties like Key and Roles.
	Page
	// Getter is the dynamic function retrieving the text value to display.
	Getter getters.Getter[string]
	// Classes represents additional CSS classes applied to the output HTML div wrapper.
	// (Discouraged: Use layout containers or theme styling instead of custom styling overrides).
	Classes string
}

// GetKey returns the unique key identifier for this FieldText component.
func (e FieldText) GetKey() string {
	return e.Key
}

// GetRoles returns the authorized roles required to view this FieldText.
func (e FieldText) GetRoles() []string {
	return e.Roles
}

// Build compiles the FieldText component into a Div Node wrapping the resolved plain text.
func (e FieldText) Build(ctx context.Context) Node {
	value := ""
	if e.Getter != nil {
		v, err := e.Getter(ctx)
		if err != nil {
			slog.Error("FieldText getter failed", "error", err, "key", e.Key)
			return ContainerError{Error: getters.Static(err)}.Build(ctx)
		}
		value = v
	}
	return Div(Class(e.Classes), Text(value))
}
