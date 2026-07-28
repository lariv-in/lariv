package components

import (
	"context"
	"html/template"
	"io"

	"github.com/lariv-in/lariv/getters"
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

// Build compiles the ContainerError component into HTML containing the children and optional error message.
func (e ContainerError) Build(cat Catalog, ctx context.Context, w io.Writer) error {
	children, err := RenderChildren(cat, ctx, e.Children)
	if err != nil {
		return err
	}
	var errorMessage string
	if e.Error != nil {
		if errVal, _ := e.Error(ctx); errVal != nil {
			errorMessage = errVal.Error()
		}
	}
	return Execute(w, "container_error", struct {
		Children     template.HTML
		ErrorMessage string
	}{Children: children, ErrorMessage: errorMessage})
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
