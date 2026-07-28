package components

import (
	"context"
	"html/template"
	"io"
)

// ButtonSubmit represents a standard submit button used within forms to post data.
// It is styled as a primary button by default.
type ButtonSubmit struct {
	// Page embeds common component properties like Key and Roles.
	Page
	// Label is the display text shown on the button.
	Label string
	// Icon is the name of an optional icon to display alongside the text.
	Icon string
	// IconClasses represents additional CSS classes applied to the Icon.
	IconClasses string
	// Classes represents additional CSS classes for the button container.
	Classes string
}

// GetKey returns the unique key identifier for this ButtonSubmit component.
func (e ButtonSubmit) GetKey() string {
	return e.Key
}

// GetRoles returns the authorized roles required to view this ButtonSubmit.
func (e ButtonSubmit) GetRoles() []string {
	return e.Roles
}

// Build compiles the ButtonSubmit component into a submit button.
func (e ButtonSubmit) Build(cat Catalog, ctx context.Context, w io.Writer) error {
	var iconHTML template.HTML
	if e.Icon != "" {
		h, err := RenderHTML(&Icon{Name: e.Icon, Classes: e.IconClasses}, cat, ctx)
		if err != nil {
			return err
		}
		iconHTML = h
	}

	classes := "btn btn-primary " + e.Classes
	if e.Icon != "" && e.Label != "" {
		classes += " inline-flex items-center gap-2"
	}

	return Execute(w, "button_submit", struct {
		Classes string
		Icon    template.HTML
		Label   string
	}{Classes: classes, Icon: iconHTML, Label: e.Label})
}
