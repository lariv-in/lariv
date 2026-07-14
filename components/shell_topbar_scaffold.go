package components

import (
	"context"

	. "maragu.dev/gomponents"
)

// ShellTopbarScaffold represents a page layout document scaffold wrapper featuring a top navigation bar.
// It nests its child components inside a [LayoutTopbar] container wrapped by the global [ShellBase] body document structure.
//
// Use Cases:
//   - Defining page templates featuring top navigation elements (menus, widgets) but omitting left-hand sidebar layout columns.
//
// Example:
//
//	 &components.ShellTopbarScaffold{
//	     Children: []components.PageInterface{
//	         &components.FieldTitle{Title: "Portal home"},
//	     },
//	 }
type ShellTopbarScaffold struct {
	// Page embeds common component properties like Key and Roles.
	Page
	// Children represents the slice of nested sub-components rendered in the topbar page content area.
	Children  []PageInterface
	// ExtraHead represents the slice of custom header tags (e.g. metadata, scripts, links) injected in the HTML head.
	ExtraHead []PageInterface
}

// Body compiles the core page content wrapper inside the parent HTML document shell structure.
func (e ShellTopbarScaffold) Body(ctx context.Context) Node {
	return ShellBase{
		ExtraHead: e.ExtraHead,
		Children: []PageInterface{
			LayoutTopbar{
				Children: e.Children,
			},
		},
	}.Body(ctx)
}

// Build compiles the ShellTopbarScaffold component into base Shell elements.
func (e ShellTopbarScaffold) Build(ctx context.Context) Node {
	return Render(ShellBase{
		ExtraHead: e.ExtraHead,
		Children: []PageInterface{
			LayoutTopbar{
				Children: e.Children,
			},
		},
	}, ctx)
}

// GetKey returns the unique key identifier for this ShellTopbarScaffold component.
func (e ShellTopbarScaffold) GetKey() string {
	return e.Key
}

// GetRoles returns the authorized roles required to view this ShellTopbarScaffold.
func (e ShellTopbarScaffold) GetRoles() []string {
	return e.Roles
}

// GetChildren returns the slice of nested sub-components.
func (e ShellTopbarScaffold) GetChildren() []PageInterface {
	return e.Children
}

// SetChildren replaces the slice of nested sub-components.
func (e *ShellTopbarScaffold) SetChildren(children []PageInterface) {
	e.Children = children
}
