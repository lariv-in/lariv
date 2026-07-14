package components

import (
	"context"

	. "maragu.dev/gomponents"
)

// ShellSimpleScaffold represents a single-column, centered layout document scaffold wrapper.
// It nests its child components inside a centered [LayoutSimple] container wrapped by the global [ShellBase] body document structure.
//
// Use Cases:
//   - Defining basic informational layouts like privacy policies, terms of service, simple utility grids, or standalone landing pages.
//
// Example:
//
//	 &components.ShellSimpleScaffold{
//	     Children: []components.PageInterface{
//	         &components.FieldTitle{Title: "Terms of Service"},
//	         &components.FieldText{Text: getters.Static("Acceptable usage policy details...")},
//	     },
//	 }
type ShellSimpleScaffold struct {
	// Page embeds common component properties like Key and Roles.
	Page
	// Children represents the slice of nested sub-components rendered in the centered single-column layout.
	Children  []PageInterface
	// ExtraHead represents the slice of custom header tags (e.g. metadata, scripts, links) injected in the HTML head.
	ExtraHead []PageInterface
}

// Body compiles the core page content wrapper inside the parent HTML document shell structure.
func (e ShellSimpleScaffold) Body(ctx context.Context) Node {
	return ShellBase{
		ExtraHead: e.ExtraHead,
		Children: []PageInterface{
			LayoutSimple{
				Children: e.Children,
			},
		},
	}.Body(ctx)
}

// Build compiles the ShellSimpleScaffold component into base Shell elements.
func (e ShellSimpleScaffold) Build(ctx context.Context) Node {
	return Render(ShellBase{
		ExtraHead: e.ExtraHead,
		Children: []PageInterface{
			LayoutSimple{
				Children: e.Children,
			},
		},
	}, ctx)
}

// GetKey returns the unique key identifier for this ShellSimpleScaffold component.
func (e ShellSimpleScaffold) GetKey() string {
	return e.Key
}

// GetRoles returns the authorized roles required to view this ShellSimpleScaffold.
func (e ShellSimpleScaffold) GetRoles() []string {
	return e.Roles
}

// GetChildren returns the slice of nested sub-components.
func (e ShellSimpleScaffold) GetChildren() []PageInterface {
	return e.Children
}

// SetChildren replaces the slice of nested sub-components.
func (e *ShellSimpleScaffold) SetChildren(children []PageInterface) {
	e.Children = children
}
