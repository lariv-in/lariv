package components

import (
	"context"
	"fmt"

	"maragu.dev/gomponents"
	"maragu.dev/gomponents/html"
)

// ContainerColumn layout container stacks its child components vertically.
// It applies a flexbox column flow ("flex flex-col gap-1") by default.
//
// Use Cases:
//   - Stacking input fields, labels, and helper text vertically within forms.
//   - Creating page sidebars or primary layout grids stacked vertically.
//
// Example:
//
//	 &components.ContainerColumn{
//	     Classes: "w-full max-w-md p-4 bg-base-100 shadow",
//	     Children: []components.PageInterface{
//	         &components.FieldText{Getter: getters.Static("Top Item")},
//	         &components.FieldText{Getter: getters.Static("Bottom Item")},
//	     },
//	 }
type ContainerColumn struct {
	// Page embeds common component properties like Key and Roles.
	Page
	// Children represents the nested components inside the column layout container.
	Children []PageInterface
	// Classes represents additional CSS classes applied to the wrapping div container.
	// (Discouraged: Limit custom class overrides to layout utilities like margins or alignment).
	Classes  string
}

// Build compiles the ContainerColumn component into a div Node containing stacked child nodes.
func (e ContainerColumn) Build(ctx context.Context) gomponents.Node {
	group := gomponents.Group{}
	for _, child := range e.Children {
		group = append(group, Render(child, ctx))
	}
	return html.Div(html.Class(fmt.Sprintf("flex flex-col gap-1 %s", e.Classes)), group)
}

// GetKey returns the unique key identifier for this ContainerColumn component.
func (e ContainerColumn) GetKey() string {
	return e.Key
}

// GetRoles returns the authorized roles required to view this ContainerColumn.
func (e ContainerColumn) GetRoles() []string {
	return e.Roles
}

// GetChildren returns the slice of child components inside the column layout container.
func (e ContainerColumn) GetChildren() []PageInterface {
	return e.Children
}

// SetChildren overwrites the child components inside the column layout container.
func (e *ContainerColumn) SetChildren(children []PageInterface) {
	e.Children = children
}
