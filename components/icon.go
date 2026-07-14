package components

import (
	"context"
	"fmt"

	"maragu.dev/gomponents"
	"maragu.dev/gomponents/html"
)

// Icon represents a graphic icon component retrieved dynamically from the Heroicons library via the Iconify design API.
// It renders an HTML span tag with class "heroicon" and sets the background SVG URL style attribute dynamically using the Name field.
//
// Use Cases:
//   - Showing semantic indicators, status markings, or navigation menu symbols (e.g., shopping carts, arrows, settings cog).
//
// Example:
//
//	 &components.Icon{
//	     Name: "check-circle",
//	 }
type Icon struct {
	// Page embeds common component properties like Key and Roles.
	Page
	// Name represents the exact name identifier of the Heroicon to load (e.g., "check-circle" or "academic-cap").
	Name    string
	// Classes represents additional CSS classes applied to the output HTML span element.
	// (Discouraged: Use layout containers or theme styling instead of custom styling overrides).
	Classes string
	// Attrs is a slice of extra HTML attributes/gomponents nodes applied to the span element.
	Attrs   []gomponents.Node
}

// GetKey returns the unique key identifier for this Icon component.
func (e Icon) GetKey() string {
	return e.Key
}

// GetRoles returns the authorized roles required to view this Icon.
func (e Icon) GetRoles() []string {
	return e.Roles
}

// Build compiles the Icon component into a Span Node displaying the Heroicon.
func (e Icon) Build(ctx context.Context) gomponents.Node {
	nodes := []gomponents.Node{
		html.Class(fmt.Sprintf("heroicon %s", e.Classes)),
		html.Style(fmt.Sprintf("--heroicon-url: url('https://api.iconify.design/heroicons/%s.svg')", e.Name)),
	}
	nodes = append(nodes, e.Attrs...)
	return html.Span(nodes...)
}
