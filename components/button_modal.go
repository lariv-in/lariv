package components

import (
	"context"
	"html/template"
	"io"

	"github.com/lariv-in/lariv/getters"
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
	// Attr is an optional Getter that yields additional HTML or HTMX attributes to attach to the button.
	Attr getters.Getter[HTMLAttributes]
}

// GetKey returns the unique key identifier for this ButtonModal component.
func (e ButtonModal) GetKey() string {
	return e.Key
}

// GetRoles returns the authorized roles required to view this ButtonModal.
func (e ButtonModal) GetRoles() []string {
	return e.Roles
}

// Build compiles the ButtonModal component into a button inside a div container.
func (e ButtonModal) Build(cat Catalog, ctx context.Context, w io.Writer) error {
	url := ""
	if e.Url != nil {
		if v, err := e.Url(ctx); err == nil {
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
		return ContainerError{Error: getters.Static(err)}.Build(cat, ctx, w)
	}

	return Execute(w, "button_modal", struct {
		URL      string
		Classes  string
		HXTarget string
		HXSwap   string
		Attrs    HTMLAttributes
		Icon     template.HTML
		Label    string
	}{
		URL:      url,
		Classes:  buttonClasses,
		HXTarget: HTMXTargetBodyModal,
		HXSwap:   HTMXSwapBodyModal,
		Attrs:    attrs,
		Icon:     iconHTML,
		Label:    e.Label,
	})
}
