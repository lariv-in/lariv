package components

import (
	"context"
	"html/template"
	"io"
)

// LayoutSidebar represents a responsive two-column dashboard shell layout component.
// It divides the page into a collapsible side navigation drawer (Sidebar) and a main content viewport (Children).
// It manages sidebar visibility and desktop/mobile responsiveness natively via Alpine.js (`showLeft`), preventing layout shifts during HTMX updates.
//
// Use Cases:
//   - Framing primary administration consoles, user dashboard panels, setting portals, or sidebar navigations.
//
// Example:
//
//	 &components.LayoutSidebar{
//	     Sidebar: []components.PageInterface{
//	         &components.Menu{...},
//	     },
//	     Children: []components.PageInterface{
//	         &components.FieldTitle{Title: "System Settings"},
//	         &components.Form[Settings]{...},
//	     },
//	 }
type LayoutSidebar struct {
	// Page embeds common component properties like Key and Roles.
	Page
	// Sidebar represents the slice of sub-components rendered in the collapsible left navigation drawer.
	Sidebar []PageInterface
	// Children represents the slice of main sub-components rendered in the viewport next to the sidebar.
	Children []PageInterface
}

// Build compiles the LayoutSidebar component into responsive sidebar and content containers.
func (e LayoutSidebar) Build(cat Catalog, ctx context.Context, w io.Writer) error {
	sidebar, err := RenderChildren(cat, ctx, e.Sidebar)
	if err != nil {
		return err
	}
	children, err := RenderChildren(cat, ctx, e.Children)
	if err != nil {
		return err
	}
	menuIcon, err := RenderHTML(Icon{Name: "bars-3"}, cat, ctx)
	if err != nil {
		return err
	}
	xData := `{
        showLeft: window.innerWidth >= 768,
        isMobile: window.innerWidth < 768,
        messages: []
}`
	return Execute(w, "layout_sidebar", struct {
		XData    string
		Sidebar  template.HTML
		MenuIcon template.HTML
		Children template.HTML
	}{XData: xData, Sidebar: sidebar, MenuIcon: menuIcon, Children: children})
}

// GetKey returns the unique key identifier for this LayoutSidebar component.
func (e LayoutSidebar) GetKey() string {
	return e.Key
}

// GetRoles returns the authorized roles required to view this LayoutSidebar.
func (e LayoutSidebar) GetRoles() []string {
	return e.Roles
}

// GetChildren aggregates and returns both the Sidebar and Children component slices.
func (e LayoutSidebar) GetChildren() []PageInterface {
	return append(e.Sidebar, e.Children...)
}

// SetChildren replaces navigation and main viewport child elements sequentially.
func (e *LayoutSidebar) SetChildren(children []PageInterface) {
	offset := 0
	nSidebar := len(e.Sidebar)
	end := min(offset+nSidebar, len(children))
	e.Sidebar = children[offset:end]
	offset = end
	if offset >= len(children) {
		return
	}
	nContent := len(e.Children)
	end = min(offset+nContent, len(children))
	e.Children = children[offset:end]
	offset = end
	if offset < len(children) {
		e.Children = append(e.Children, children[offset:]...)
	}
}
