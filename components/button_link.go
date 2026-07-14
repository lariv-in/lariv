package components

import (
	"context"

	"github.com/lariv-in/lago/getters"
	"maragu.dev/gomponents"
	"maragu.dev/gomponents/html"
)

// ButtonLink represents an anchor link styled as a button.
// It supports dynamic or static labels (via getters) and targets.
type ButtonLink struct {
	// Page embeds common component properties like Key and Roles.
	Page
	// Label is a Getter resolving to the button's display text.
	Label getters.Getter[string]
	// Link is a Getter that resolves to the destination URL (href) of the link.
	Link getters.Getter[string]
	// Icon is the name of an optional icon to display alongside the text.
	Icon string
	// IconClasses represents additional CSS classes applied to the Icon.
	IconClasses string
	// Classes represents additional CSS classes for the button container.
	Classes string
}

// GetKey returns the unique key identifier for this ButtonLink component.
func (e ButtonLink) GetKey() string {
	return e.Key
}

// GetRoles returns the authorized roles required to view this ButtonLink.
func (e ButtonLink) GetRoles() []string {
	return e.Roles
}

// Build compiles the ButtonLink component into a gomponents Node representing an HTML <a> element.
func (e ButtonLink) Build(ctx context.Context) gomponents.Node {
	link := ""
	if e.Link != nil {
		if v, err := e.Link(ctx); err == nil {
			link = v
		}
	}
	label := ""
	if e.Label != nil {
		if v, err := e.Label(ctx); err == nil {
			label = v
		}
	}

	content := gomponents.Group{}
	if e.Icon != "" {
		content = append(content, Render(Icon{Name: e.Icon, Classes: e.IconClasses}, ctx))
	}
	if label != "" {
		content = append(content, gomponents.Text(label))
	}

	classes := "btn " + e.Classes
	if e.Icon != "" && label != "" {
		classes += " flex items-center gap-2"
	}
	return html.A(html.Href(link), html.Class(classes), content)
}
