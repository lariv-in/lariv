package components

import (
	"context"
	"log/slog"

	"github.com/lariv-in/lariv/getters"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

// FieldLink represents a read-only hyperlink field.
// It renders an HTML anchor <a> tag with href and display label resolved dynamically from context.
// If Href evaluates to an empty string, the component falls back to rendering a plain text <div>.
//
// Use Cases:
//   - Linking to entity detail pages from lists or summaries (e.g., clicking on a customer's name to view their profile).
//   - Displaying clickable reference URLs or external links.
//
// Example:
//
//	&components.FieldLink{
//	    Href:    lariv.RoutePath("users.UserDetail", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("$in.UserID"))}),
//	    Label:   getters.Key[string]("$in.User.Name"),
//	    Classes: "link link-primary",
//	}
type FieldLink struct {
	// Page embeds common component properties like Key and Roles.
	Page
	// Href is a Getter resolving to the target URL for the anchor link.
	Href getters.Getter[string]
	// Label is an optional Getter resolving to the display text. If empty or not set, the resolved Href value is used.
	Label getters.Getter[string]
	// Classes represents additional CSS classes applied to the output HTML element.
	// (Discouraged: Use layout containers or theme styling instead of custom styling overrides).
	Classes string
	// Attr is an optional Getter returning additional HTML attributes to apply to the link.
	Attr getters.Getter[Node]
}

// GetKey returns the unique key identifier for this FieldLink component.
func (e FieldLink) GetKey() string {
	return e.Key
}

// GetRoles returns the authorized roles required to view this FieldLink.
func (e FieldLink) GetRoles() []string {
	return e.Roles
}

// Build compiles the FieldLink component into an anchor Node, or a text Div if the URL is empty.
func (e FieldLink) Build(ctx context.Context) Node {
	href := ""
	if e.Href != nil {
		v, err := e.Href(ctx)
		if err != nil {
			slog.Error("FieldLink href getter failed", "error", err, "key", e.Key)
			return ContainerError{Error: getters.Static(err)}.Build(ctx)
		}
		href = v
	}
	label := href
	if e.Label != nil {
		if v, err := e.Label(ctx); err == nil && v != "" {
			label = v
		}
	}
	if href == "" {
		return Div(Class(e.Classes), Text(label))
	}
	var extra Node = Raw("")
	if e.Attr != nil {
		if n, err := e.Attr(ctx); err == nil && n != nil {
			extra = n
		}
	}
	return A(Href(href), Class(e.Classes), extra, Text(label))
}
