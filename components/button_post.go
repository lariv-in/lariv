package components

import (
	"context"
	"html/template"
	"io"
	"log/slog"

	"github.com/lariv-in/lariv/getters"
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
	// Attr is an optional Getter yielding additional HTML/HTMX attributes to merge onto the submit button.
	Attr getters.Getter[HTMLAttributes]
}

// GetKey returns the unique key identifier for this ButtonPost component.
func (e ButtonPost) GetKey() string {
	return e.Key
}

// GetRoles returns the authorized roles required to view this ButtonPost.
func (e ButtonPost) GetRoles() []string {
	return e.Roles
}

// Build compiles the ButtonPost component into a submit button inside a POST form.
func (e ButtonPost) Build(cat Catalog, ctx context.Context, w io.Writer) error {
	url := ""
	if e.URL != nil {
		if v, err := e.URL(ctx); err == nil {
			url = v
		}
	}

	var iconHTML template.HTML
	if e.Icon != "" {
		h, err := RenderHTML(&Icon{Name: e.Icon, Classes: e.IconClasses}, cat, ctx)
		if err != nil {
			return err
		}
		iconHTML = h
	}

	buttonClasses := "btn " + e.Classes
	if e.Icon != "" && e.Label != "" {
		buttonClasses += " inline-flex items-center gap-2"
	}

	attrs, err := ResolveAttrs(ctx, e.Attr)
	if err != nil {
		slog.Error("ButtonPost Attr getter failed", "error", err, "key", e.Key)
		attrs = HTMLAttributes{}
	}

	return Execute(w, "button_post", struct {
		URL     string
		Classes string
		Attrs   HTMLAttributes
		Icon    template.HTML
		Label   string
	}{URL: url, Classes: buttonClasses, Attrs: attrs, Icon: iconHTML, Label: e.Label})
}
