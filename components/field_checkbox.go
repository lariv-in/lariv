package components

import (
	"context"
	"log/slog"

	"github.com/lariv-in/lago/getters"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

// FieldCheckbox represents a read-only field that displays boolean values as icons.
// Renders a green "check-circle" icon for true values, and a red "x-circle" icon for false values.
//
// Use Cases:
//   - Showing status flags in database entity detail views (e.g. "Is Active", "Email Verified", "Is Admin").
//
// Example:
//
//	&components.FieldCheckbox{
//	    Getter: getters.Key[bool]("$in.IsVerified"),
//	}
type FieldCheckbox struct {
	// Page embeds common component properties like Key and Roles.
	Page
	// Getter is the dynamic function retrieving the boolean state value.
	Getter getters.Getter[bool]
}

// GetKey returns the unique key identifier for this FieldCheckbox component.
func (e FieldCheckbox) GetKey() string {
	return e.Key
}

// GetRoles returns the authorized roles required to view this FieldCheckbox.
func (e FieldCheckbox) GetRoles() []string {
	return e.Roles
}

// Build compiles the FieldCheckbox component into a Span Node containing the check or X icon.
func (e FieldCheckbox) Build(ctx context.Context) Node {
	truthy := false
	if e.Getter != nil {
		value, err := e.Getter(ctx)
		if err != nil {
			slog.Error("FieldCheckbox getter failed", "error", err, "key", e.Key)
			return ContainerError{Error: getters.Static(err)}.Build(ctx)
		}
		truthy = value
	}

	if truthy {
		return Span(Render(Icon{Name: "check-circle", Classes: "text-success"}, ctx))
	}
	return Span(Render(Icon{Name: "x-circle", Classes: "text-error"}, ctx))
}
