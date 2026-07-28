package components

import (
	"context"
	"html/template"
	"io"
)

// ClientShow toggles the visibility of its children using Alpine.js's x-show directive.
// Unlike [ClientIf], it keeps the children in the DOM and toggles their CSS display style.
// It must be placed within a parent Alpine.js scope (such as [ClientData]) to read the reactive condition.
//
// Use Cases:
//   - Showing or hiding tooltips, dropdown menus, or warning panels interactively on the client.
//   - Quick toggling of content sections where keeping elements in the DOM preserves state.
//
// Example:
//
//	 &components.ClientData{
//	     Data: "{ showMore: false }",
//	     Children: []components.PageInterface{
//	         components.Button{Label: "Toggle info", Attr: Attr("@click", "showMore = !showMore")},
//	         &components.ClientShow{
//	             Condition: "showMore",
//	             Children: []components.PageInterface{
//	                 &components.FieldText{Getter: getters.Static("Extended details...")},
//	             },
//	         },
//	     },
//	 }
type ClientShow struct {
	// Page embeds common component properties like Key and Roles.
	Page
	// Condition specifies a valid client-side Alpine.js expression to evaluate (e.g. "showMore").
	Condition string
	// Children represents the child components whose visibility is toggled.
	Children []PageInterface
}

// Build compiles the ClientShow component into a div containing Alpine x-show attribute.
func (e ClientShow) Build(cat Catalog, ctx context.Context, w io.Writer) error {
	children, err := RenderChildren(cat, ctx, e.Children)
	if err != nil {
		return err
	}
	return Execute(w, "client_show", struct {
		Condition string
		Children  template.HTML
	}{Condition: e.Condition, Children: children})
}

// GetKey returns the unique key identifier for this ClientShow component.
func (e ClientShow) GetKey() string {
	return e.Key
}

// GetRoles returns the authorized roles required to view this ClientShow.
func (e ClientShow) GetRoles() []string {
	return e.Roles
}

// GetChildren returns the slice of child components managed by this ClientShow wrapper.
func (e ClientShow) GetChildren() []PageInterface {
	return e.Children
}

// SetChildren overwrites the child components managed by this ClientShow wrapper.
func (e *ClientShow) SetChildren(children []PageInterface) {
	e.Children = children
}
