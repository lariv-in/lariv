package components

import (
	"context"
	"html/template"
	"io"
)

// LayoutCard represents a page layout wrapper component designed to center its children inside a standalone card element.
// Layout components are special structural nodes in Lariv; they establish the main page-level grid or shell wrapper,
// framing child page components inside clean canvas containers (in this case, a centered card overlay on base-200 background).
// It also integrates a global loading progress indicator (#global-loading-indicator) pinned at the top.
//
// Use Cases:
//   - Framing user login, password recovery, registration, or simple confirmation screens.
//
// Example:
//
//	&components.LayoutCard{
//	    Children: []components.PageInterface{
//	        &components.FieldTitle{Title: "Login"},
//	        &components.Form[LoginParams]{...},
//	    },
//	}
type LayoutCard struct {
	// Page embeds common component properties like Key and Roles.
	Page
	// Children represents the slice of nested sub-components rendered inside the card body.
	Children []PageInterface
}

// Build compiles the LayoutCard component into a centered card overlay with progress bars.
func (e LayoutCard) Build(cat Catalog, ctx context.Context, w io.Writer) error {
	children, err := RenderChildren(cat, ctx, e.Children)
	if err != nil {
		return err
	}
	return Execute(w, "layout_card", struct {
		Children template.HTML
	}{Children: children})
}

// GetKey returns the unique key identifier for this LayoutCard component.
func (e LayoutCard) GetKey() string {
	return e.Key
}

// GetRoles returns the authorized roles required to view this LayoutCard.
func (e LayoutCard) GetRoles() []string {
	return e.Roles
}

// GetChildren returns the slice of nested sub-components.
func (e LayoutCard) GetChildren() []PageInterface {
	return e.Children
}

// SetChildren replaces the slice of nested sub-components.
func (e *LayoutCard) SetChildren(children []PageInterface) {
	e.Children = children
}
