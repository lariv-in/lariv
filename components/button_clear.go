package components

import (
	"context"
	"html/template"
	"io"
)

// ButtonClear represents a button that clears/resets all inputs in its containing form.
// It uses a simple inline JavaScript onclick handler to clear the values of inputs,
// selects, and textareas.
type ButtonClear struct {
	// Page embeds common component properties like Key and Roles.
	Page
	// Label is the display text on the button. Defaults to "Clear" if empty.
	Label string
	// Icon is the name of an optional icon to display alongside the text.
	Icon string
	// IconClasses represents additional CSS classes applied to the Icon.
	IconClasses string
	// Classes represents additional CSS classes for the button container.
	Classes string
}

// GetKey returns the unique key identifier for this ButtonClear component.
func (e ButtonClear) GetKey() string {
	return e.Key
}

// GetRoles returns the authorized roles required to view this ButtonClear.
func (e ButtonClear) GetRoles() []string {
	return e.Roles
}

// Build compiles the ButtonClear component into HTML.
func (e ButtonClear) Build(cat Catalog, ctx context.Context, w io.Writer) error {
	label := e.Label
	if label == "" {
		label = "Clear"
	}

	var iconHTML template.HTML
	if e.Icon != "" {
		h, err := RenderHTML(&Icon{Name: e.Icon, Classes: e.IconClasses}, cat, ctx)
		if err != nil {
			return err
		}
		iconHTML = h
	}

	classes := "btn btn-ghost my-2 " + e.Classes
	if e.Icon != "" && label != "" {
		classes += " inline-flex items-center gap-2"
	}

	return Execute(w, "button_clear", struct {
		Classes string
		Icon    template.HTML
		Label   string
	}{Classes: classes, Icon: iconHTML, Label: label})
}
