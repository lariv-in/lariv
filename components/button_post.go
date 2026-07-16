package components

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/lariv-in/lariv/getters"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

// ButtonPost represents a button that performs a POST request via an HTMX-boosted form.
// This is useful for triggering state changes or actions (e.g., initiating background tasks,
// cancellations, regenerations) without full page reloads.
type ButtonPost struct {
	// Page embeds common component properties like Key and Roles.
	Page
	// Label is the display text shown on the button.
	Label string
	// URL is a Getter that resolves the form action target URL for the POST request.
	URL getters.Getter[string]
	// Icon is the name of an optional icon to display alongside the text.
	Icon string
	// IconClasses represents additional CSS classes applied to the Icon.
	IconClasses string
	// Classes represents additional CSS classes for the button container.
	Classes string
	// Attr is an optional Getter yielding additional HTML/HTMX attributes (Node) to merge onto the submit button.
	Attr getters.Getter[Node]
}

// GetKey returns the unique key identifier for this ButtonPost component.
func (e ButtonPost) GetKey() string {
	return e.Key
}

// GetRoles returns the authorized roles required to view this ButtonPost.
func (e ButtonPost) GetRoles() []string {
	return e.Roles
}

// Build compiles the ButtonPost component into a gomponents Node representing a submit button inside a POST form.
func (e ButtonPost) Build(ctx context.Context) Node {
	url := ""
	if e.URL != nil {
		if v, err := e.URL(ctx); err == nil {
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

	return Form(
		Action(url), Method(http.MethodPost),
		// Use htmx boost so the POST is handled via HTMX without a
		// full-page navigation; the response (e.g. updated detail view
		// showing "Generating..." state) will be swapped in-place.
		Attr("hx-boost", "true"),
		Attr("@click.stop", ""),
		Button(
			Type("submit"),
			Class(buttonClasses),
			Iff(e.Attr != nil, func() Node {
				n, err := e.Attr(ctx)
				if err != nil {
					slog.Error("ButtonPost Attr getter failed", "error", err, "key", e.Key)
					return Raw("")
				}
				if n == nil {
					return Raw("")
				}
				return n
			}),
			content,
		),
	)
}
