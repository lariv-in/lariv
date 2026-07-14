package components

import (
	"context"

	"maragu.dev/gomponents"
)

// ContainerHTML layout container wraps child components in a custom HTML layout function.
// This allows developers to wrap nested pages in arbitrary custom HTML tags or Tailwind cards dynamically.
//
// Use Cases:
//   - Wrapping input panels inside custom Tailwind/DaisyUI cards.
//   - Enclosing list items or fields in native detail disclosures, fieldsets, or form tags.
//
// Example:
//
//	 &components.ContainerHTML{
//	     Children: []components.PageInterface{
//	         &components.FieldText{Getter: getters.Static("Inner content")},
//	     },
//	     HTML: func(ctx context.Context, children gomponents.Node) gomponents.Node {
//	         return html.Div(html.Class("card bg-base-200 p-4 shadow"), children)
//	     },
//	 }
type ContainerHTML struct {
	// Page embeds common component properties like Key and Roles.
	Page
	// Children represents the nested components to render inside this wrapper.
	Children []PageInterface
	// HTML represents the custom function that receives the rendered children nodes and returns the wrapped HTML structure.
	HTML     func(context.Context, gomponents.Node) gomponents.Node
}

// Build compiles the ContainerHTML component by rendering children and executing the HTML layout callback.
func (e ContainerHTML) Build(ctx context.Context) gomponents.Node {
	group := gomponents.Group{}
	for _, child := range e.Children {
		group = append(group, Render(child, ctx))
	}
	if e.HTML != nil {
		return e.HTML(ctx, group)
	}
	return group
}

// GetKey returns the unique key identifier for this ContainerHTML component.
func (e ContainerHTML) GetKey() string {
	return e.Key
}

// GetRoles returns the authorized roles required to view this ContainerHTML.
func (e ContainerHTML) GetRoles() []string {
	return e.Roles
}

// GetChildren returns the slice of child components inside this ContainerHTML wrapper.
func (e ContainerHTML) GetChildren() []PageInterface {
	return e.Children
}

// SetChildren overwrites the child components inside this ContainerHTML wrapper.
func (e *ContainerHTML) SetChildren(children []PageInterface) {
	e.Children = children
}
