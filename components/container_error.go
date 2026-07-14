package components

import (
	"context"

	"github.com/lariv-in/lago/getters"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

// ContainerError wraps child components and conditionally appends a visible error label.
// If the Error getter resolves to a non-nil error, the error message is displayed as a red helper text beneath the children.
//
// Use Cases:
//   - Showing top-level global validation errors on form submissions.
//   - Surrounding detail sections or data grids to catch and display dynamic retrieval errors.
//
// Example:
//
//	&components.ContainerError{
//	    Error: getters.Key[error]("$error._global"),
//	    Children: []components.PageInterface{
//	        &components.FieldText{Getter: getters.Static("Form Contents")},
//	    },
//	}
type ContainerError struct {
	// Page embeds common component properties like Key and Roles.
	Page
	// Children represents the nested components inside the error container.
	Children []PageInterface
	// Error is a Getter that resolves the error state. If non-nil, its message is displayed below the children.
	Error getters.Getter[error]
}

// Build compiles the ContainerError component into a div Node containing the child nodes and the optional error message label.
func (e ContainerError) Build(ctx context.Context) Node {
	group := Group{}
	for _, child := range e.Children {
		group = append(group, Render(child, ctx))
	}

	var errorNode Node
	if e.Error != nil {
		err, _ := e.Error(ctx)
		if err != nil {
			errorNode = Span(Class("text-sm text-error"), Text(err.Error()))
		}
	}

	return Div(Class("flex flex-col gap-1 w-full"), group, errorNode)
}

// GetKey returns the unique key identifier for this ContainerError component.
func (e ContainerError) GetKey() string {
	return e.Key
}

// GetRoles returns the authorized roles required to view this ContainerError.
func (e ContainerError) GetRoles() []string {
	return e.Roles
}

// GetChildren returns the slice of child components inside this ContainerError wrapper.
func (e ContainerError) GetChildren() []PageInterface {
	return e.Children
}

// SetChildren overwrites the child components inside this ContainerError wrapper.
func (e *ContainerError) SetChildren(children []PageInterface) {
	e.Children = children
}
