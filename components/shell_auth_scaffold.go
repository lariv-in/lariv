package components

import (
	"context"

	. "maragu.dev/gomponents"
)

// ShellAuthScaffold represents the global base page shell scaffold for authentication views.
// It wraps its children inside a centered [LayoutCard] container and renders them inside the standard [ShellBase] HTML body structure.
//
// Use Cases:
//   - Defining HTML document wrappers for login pages, user signup flows, or credentials reset screens.
//
// Example:
//
//	 &components.ShellAuthScaffold{
//	     Children: []components.PageInterface{
//	         &components.FieldTitle{Title: "Sign In"},
//	         &components.Form[LoginParams]{...},
//	     },
//	 }
type ShellAuthScaffold struct {
	// Page embeds common component properties like Key and Roles.
	Page
	// Children represents the slice of nested sub-components rendered in the centered card.
	Children  []PageInterface
	// ExtraHead represents the slice of custom header tags (e.g. meta, scripts, links) injected in the HTML head.
	ExtraHead []PageInterface
}

// Body compiles the core page content wrapper inside the parent HTML document shell structure.
func (e ShellAuthScaffold) Body(ctx context.Context) Node {
	return ShellBase{
		ExtraHead: e.ExtraHead,
		Children:  []PageInterface{LayoutCard{Page{}, e.Children}},
	}.Body(ctx)
}

// Build compiles the ShellAuthScaffold component into base Shell elements.
func (e ShellAuthScaffold) Build(ctx context.Context) Node {
	return Render(ShellBase{
		ExtraHead: e.ExtraHead,
		Children:  []PageInterface{LayoutCard{Page{}, e.Children}},
	}, ctx)
}

// GetKey returns the unique key identifier for this ShellAuthScaffold component.
func (e ShellAuthScaffold) GetKey() string {
	return e.Key
}

// GetRoles returns the authorized roles required to view this ShellAuthScaffold.
func (e ShellAuthScaffold) GetRoles() []string {
	return e.Roles
}

// GetChildren returns the slice of nested sub-components.
func (e ShellAuthScaffold) GetChildren() []PageInterface {
	return e.Children
}

// SetChildren replaces the slice of nested sub-components.
func (e *ShellAuthScaffold) SetChildren(children []PageInterface) {
	e.Children = children
}
