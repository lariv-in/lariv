package components

import (
	"context"
	"fmt"

	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

// ContainerRow layout container arranges its child components horizontally.
// It applies a flexbox row flow ("flex flex-row gap-1") by default.
//
// Use Cases:
//   - Aligning actions or buttons side-by-side (e.g., placing Save and Cancel buttons in a form footer).
//   - Arranging tags, filters, badges, or small inputs horizontally.
//
// Example:
//
//	 &components.ContainerRow{
//	     Classes: "justify-end gap-2 mt-4",
//	     Children: []components.PageInterface{
//	         &components.ButtonClear{Label: "Cancel"},
//	         &components.ButtonSubmit{Label: "Save"},
//	     },
//	 }
type ContainerRow struct {
	// Page embeds common component properties like Key and Roles.
	Page
	// Children represents the nested components inside the row layout container.
	Children []PageInterface
	// Classes represents additional CSS classes applied to the wrapping div container.
	// (Discouraged: Limit custom class overrides to layout utilities like margins or alignment).
	Classes  string
}

// Build compiles the ContainerRow component into a div Node containing horizontally laid out child nodes.
func (e ContainerRow) Build(ctx context.Context) Node {
	group := Group{}
	for _, child := range e.Children {
		group = append(group, Render(child, ctx))
	}
	return Div(Class(fmt.Sprintf("flex flex-row gap-1 %s", e.Classes)), group)
}

// GetKey returns the unique key identifier for this ContainerRow component.
func (e ContainerRow) GetKey() string {
	return e.Key
}

// GetRoles returns the authorized roles required to view this ContainerRow.
func (e ContainerRow) GetRoles() []string {
	return e.Roles
}

// GetChildren returns the slice of child components inside this row layout container.
func (e ContainerRow) GetChildren() []PageInterface {
	return e.Children
}

// SetChildren overwrites the child components inside this row layout container.
func (e *ContainerRow) SetChildren(children []PageInterface) {
	e.Children = children
}
