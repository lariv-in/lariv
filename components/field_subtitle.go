package components

import (
	"context"
	"log/slog"

	"github.com/lariv-in/lariv/getters"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

// FieldSubtitle represents a read-only subtitle or descriptor header field.
// It displays a secondary text string styled in small gray letters ("text-md text-gray-500").
//
// Use Cases:
//   - Showing subheadings, category descriptions, or description texts beneath main title fields.
//
// Example:
//
//	&components.FieldSubtitle{
//	    Getter: getters.Static("Manage your organization preferences here."),
//	}
type FieldSubtitle struct {
	// Page embeds common component properties like Key and Roles.
	Page
	// Getter is the dynamic function retrieving the subtitle string.
	Getter getters.Getter[string]
}

// GetKey returns the unique key identifier for this FieldSubtitle component.
func (e FieldSubtitle) GetKey() string {
	return e.Key
}

// GetRoles returns the authorized roles required to view this FieldSubtitle.
func (e FieldSubtitle) GetRoles() []string {
	return e.Roles
}

// Build compiles the FieldSubtitle component into a Div Node styled as gray secondary text.
func (e FieldSubtitle) Build(ctx context.Context) Node {
	value := ""
	if e.Getter != nil {
		v, err := e.Getter(ctx)
		if err != nil {
			slog.Error("FieldSubtitle getter failed", "error", err, "key", e.Key)
			return ContainerError{Error: getters.Static(err)}.Build(ctx)
		}
		value = v
	}
	return Div(Class("text-md text-gray-500"), Text(value))
}
