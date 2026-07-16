package components

import (
	"context"

	"github.com/lariv-in/lariv/getters"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

// ButtonModal represents a button that fetches and displays modal content via HTMX.
// When clicked, it makes a GET request (hx-get) to fetch modal markup and renders it
// in the page's shared modal slot.
type ButtonModal struct {
	// Page embeds common component properties like Key and Roles.
	Page
	// Label is the display text shown on the button.
	Label string
	// Url is a Getter that resolves the API endpoint URL from which the modal content is fetched.
	Url getters.Getter[string]
	// Icon is the name of an optional icon to display alongside the text.
	Icon string
	// IconClasses represents additional CSS classes applied to the Icon.
	IconClasses string
	// Classes represents additional CSS classes for the button container.
	Classes string
	// Attr is an optional Getter that yields additional HTML or HTMX attributes (Node) to attach to the button.
	Attr getters.Getter[Node]
}

// GetKey returns the unique key identifier for this ButtonModal component.
func (e ButtonModal) GetKey() string {
	return e.Key
}

// GetRoles returns the authorized roles required to view this ButtonModal.
func (e ButtonModal) GetRoles() []string {
	return e.Roles
}

// Build compiles the ButtonModal component into a gomponents Node representing a button inside a div container.
func (e ButtonModal) Build(ctx context.Context) Node {
	url := ""
	if e.Url != nil {
		if v, err := e.Url(ctx); err == nil {
			url = v
		}
	}

	content := Group{}
	if e.Icon != "" {
		content = append(content, Render(Icon{Name: e.Icon, Classes: e.IconClasses}, ctx))
	}
	if e.Label != "" {
		content = append(content, Text(e.Label))
	}

	buttonClasses := "btn " + e.Classes
	if e.Icon != "" && e.Label != "" {
		buttonClasses += " inline-flex items-center gap-2"
	}

	buttonAttrs := []Node{
		Type("button"),
		Class(buttonClasses),
		Attr("hx-get", url),
		Attr("hx-target", HTMXTargetBodyModal),
		Attr("hx-swap", HTMXSwapBodyModal),
		Attr("hx-push-url", "false"),
	}
	if e.Attr != nil {
		extra, err := e.Attr(ctx)
		if err != nil {
			return ContainerError{Error: getters.Static(err)}.Build(ctx)
		}
		if extra != nil {
			buttonAttrs = append(buttonAttrs, extra)
		}
	}
	buttonAttrs = append(buttonAttrs, content)

	return Div(
		Class("w-full fk-modal-host"),
		Button(Group(buttonAttrs)),
	)
}
