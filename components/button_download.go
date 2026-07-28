package components

import (
	"context"
	"html/template"
	"io"

	"github.com/lariv-in/lariv/getters"
)

// ButtonDownload represents an anchor link styled as a button that triggers a file download.
// It sets the 'download' attribute and disables HTMX boosting on the link to ensure standard
// browser download behavior.
type ButtonDownload struct {
	// Page embeds common component properties like Key and Roles.
	Page
	// Label is the display text on the button.
	Label string
	// Link is a Getter that resolves the URL/href of the file to be downloaded.
	Link getters.Getter[string]
	// Icon is the name of an optional icon to display alongside the text.
	Icon string
	// IconClasses represents additional CSS classes applied to the Icon.
	IconClasses string
	// Classes represents additional CSS classes for the button container.
	Classes string
}

// GetKey returns the unique key identifier for this ButtonDownload component.
func (e ButtonDownload) GetKey() string {
	return e.Key
}

// GetRoles returns the authorized roles required to view this ButtonDownload.
func (e ButtonDownload) GetRoles() []string {
	return e.Roles
}

// Build compiles the ButtonDownload component into an HTML <a> element.
func (e ButtonDownload) Build(cat Catalog, ctx context.Context, w io.Writer) error {
	link := ""
	if e.Link != nil {
		if v, err := e.Link(ctx); err == nil {
			link = v
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

	classes := "btn " + e.Classes
	if e.Icon != "" && e.Label != "" {
		classes += " inline-flex gap-2"
	}

	return Execute(w, "button_download", struct {
		Link    string
		Classes string
		Icon    template.HTML
		Label   string
	}{Link: link, Classes: classes, Icon: iconHTML, Label: e.Label})
}
