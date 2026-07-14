package components

import (
	"context"

	. "maragu.dev/gomponents"
)

// ShellScaffold represents the main authenticated dashboard page scaffolding layout.
// It structures the page by nesting [LayoutSidebar] inside [LayoutTopbar], which is in turn wrapped inside [ShellBase] HTML document body layouts.
//
// Use Cases:
//   - Defining page templates featuring top navigation bars, sidebar categories menus, and dashboard content panes.
//
// Example:
//
//	 &components.ShellScaffold{
//	     Sidebar: []components.PageInterface{
//	         &components.SidebarMenu{...},
//	     },
//	     Children: []components.PageInterface{
//	         &components.FieldTitle{Title: "Dashboard"},
//	     },
//	 }
type ShellScaffold struct {
	// Page embeds common component properties like Key and Roles.
	Page
	// Sidebar represents the slice of sub-components rendered in the left-hand navigation column.
	Sidebar   []PageInterface
	// Children represents the slice of sub-components rendered in the main dashboard content body.
	Children  []PageInterface
	// ExtraHead represents the slice of custom header tags (e.g. metadata, scripts, links) injected in the HTML head.
	ExtraHead []PageInterface
}

// Body compiles the core page content wrapper inside the parent HTML document shell structure.
func (e ShellScaffold) Body(ctx context.Context) Node {
	return ShellBase{
		ExtraHead: e.ExtraHead,
		Children: []PageInterface{
			LayoutTopbar{
				Children: []PageInterface{
					LayoutSidebar{
						Sidebar:  e.Sidebar,
						Children: e.Children,
					},
				},
			},
		},
	}.Body(ctx)
}

// Build compiles the ShellScaffold component into base Shell elements.
func (e ShellScaffold) Build(ctx context.Context) Node {
	return Render(ShellBase{
		ExtraHead: e.ExtraHead,
		Children: []PageInterface{
			LayoutTopbar{
				Children: []PageInterface{
					LayoutSidebar{
						Sidebar:  e.Sidebar,
						Children: e.Children,
					},
				},
			},
		},
	}, ctx)
}

// GetKey returns the unique key identifier for this ShellScaffold component.
func (e ShellScaffold) GetKey() string {
	return e.Key
}

// GetRoles returns the authorized roles required to view this ShellScaffold.
func (e ShellScaffold) GetRoles() []string {
	return e.Roles
}

// GetChildren returns the slice of nested sub-components.
func (e ShellScaffold) GetChildren() []PageInterface {
	return append(e.Sidebar, e.Children...)
}

// SetChildren replaces the slice of nested sub-components.
func (e *ShellScaffold) SetChildren(children []PageInterface) {
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
