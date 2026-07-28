package components

import (
	"context"
	"io"
	"log/slog"

	"github.com/lariv-in/lariv/getters"
)

// FieldTextArea represents a read-only multi-line text display field.
// It resolves a string value dynamically from context and renders it inside a div decorated with the CSS class "whitespace-pre-wrap" to preserve formatting and newlines.
//
// Use Cases:
//   - Showing long descriptions, logs, address records, or administrator notes where paragraph breaks and spaces must be preserved.
//
// Example:
//
//	&components.FieldTextArea{
//	    Getter:  getters.Key[string]("$in.AdminNotes"),
//	    Classes: "bg-base-200 border p-2 rounded",
//	}
type FieldTextArea struct {
	// Page embeds common component properties like Key and Roles.
	Page
	// Getter is the dynamic function retrieving the multi-line text string.
	Getter getters.Getter[string]
	// Classes represents additional CSS classes applied to the output HTML div wrapper.
	// (Discouraged: Use layout containers or theme styling instead of custom styling overrides).
	Classes string
}

// GetKey returns the unique key identifier for this FieldTextArea component.
func (e FieldTextArea) GetKey() string {
	return e.Key
}

// GetRoles returns the authorized roles required to view this FieldTextArea.
func (e FieldTextArea) GetRoles() []string {
	return e.Roles
}

// Build compiles the FieldTextArea component into a Div with formatting preservation classes.
func (e FieldTextArea) Build(cat Catalog, ctx context.Context, w io.Writer) error {
	value := ""
	if e.Getter != nil {
		v, err := e.Getter(ctx)
		if err != nil {
			slog.Error("FieldTextArea getter failed", "error", err, "key", e.Key)
			return ContainerError{Error: getters.Static(err)}.Build(cat, ctx, w)
		}
		value = v
	}
	return Execute(w, "field_text_area", struct {
		Classes string
		Value   string
	}{Classes: e.Classes + " whitespace-pre-wrap", Value: value})
}
