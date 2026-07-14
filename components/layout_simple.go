package components

import (
	"context"

	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

// LayoutSimple represents a clean, single-column fullscreen canvas layout component.
// Layout components are structural wrapper components; LayoutSimple acts as a basic viewport framing children inside a scrollable, padded canvas container.
// It renders the nested sub-components inside a simple scrollable Div block helper.
//
// Use Cases:
//   - Displaying plain document views (e.g. privacy policies, terms of service), simple statistics layouts, or full-width analytics dashboards.
//
// Example:
//
//	 &components.LayoutSimple{
//	     Children: []components.PageInterface{
//	         &components.FieldTitle{Title: "Privacy Policy"},
//	         &components.FieldMarkdown{Getter: getters.Static("# Title...")},
//	     },
//	 }
type LayoutSimple struct {
	// Page embeds common component properties like Key and Roles.
	Page
	// Children represents the slice of nested sub-components rendered in the viewport canvas.
	Children []PageInterface
}

// Build compiles the LayoutSimple component into a single-column padded viewport Node.
func (e LayoutSimple) Build(ctx context.Context) Node {
	return Render(ContainerHTML{
		Children: e.Children,
		HTML: func(ctx context.Context, children Node) Node {
			return Div(Class("size-full overflow-y-auto p-4"),
				children,
			)
		},
	}, ctx)
}

// GetKey returns the unique key identifier for this LayoutSimple component.
func (e LayoutSimple) GetKey() string {
	return e.Key
}

// GetRoles returns the authorized roles required to view this LayoutSimple.
func (e LayoutSimple) GetRoles() []string {
	return e.Roles
}

// GetChildren returns the slice of nested sub-components.
func (e LayoutSimple) GetChildren() []PageInterface {
	return e.Children
}

// SetChildren replaces the slice of nested sub-components.
func (e *LayoutSimple) SetChildren(children []PageInterface) {
	e.Children = children
}
