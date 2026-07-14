package components

import (
	"context"

	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

// ButtonSubmit represents a standard submit button used within forms to post data.
// It is styled as a primary button by default.
type ButtonSubmit struct {
	// Page embeds common component properties like Key and Roles.
	Page
	// Label is the display text shown on the button.
	Label       string
	// Icon is the name of an optional icon to display alongside the text.
	Icon        string
	// IconClasses represents additional CSS classes applied to the Icon.
	IconClasses string
	// Classes represents additional CSS classes for the button container.
	Classes     string
}

// GetKey returns the unique key identifier for this ButtonSubmit component.
func (e ButtonSubmit) GetKey() string {
	return e.Key
}

// GetRoles returns the authorized roles required to view this ButtonSubmit.
func (e ButtonSubmit) GetRoles() []string {
	return e.Roles
}

// Build compiles the ButtonSubmit component into a gomponents Node representing a submit button.
func (e ButtonSubmit) Build(ctx context.Context) Node {
	content := Group{}
	if e.Icon != "" {
		content = append(content, Render(Icon{Name: e.Icon, Classes: e.IconClasses}, ctx))
	}
	if e.Label != "" {
		content = append(content, Text(e.Label))
	}

	classes := "btn btn-primary " + e.Classes
	if e.Icon != "" && e.Label != "" {
		classes += " inline-flex items-center gap-2"
	}

	return Button(Type("submit"), Class(classes), content)
}
