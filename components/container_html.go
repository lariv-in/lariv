package components

import (
	"context"
	"html/template"
	"io"
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
//	     HTML: func(ctx context.Context, w io.Writer, children template.HTML) error {
//	         _, err := io.WriteString(w, `<div class="card bg-base-200 p-4 shadow">`+string(children)+`</div>`)
//	         return err
//	     },
//	 }
type ContainerHTML struct {
	// Page embeds common component properties like Key and Roles.
	Page
	// Children represents the nested components to render inside this wrapper.
	Children []PageInterface
	// HTML represents the custom function that receives the rendered children HTML and writes the wrapped structure.
	HTML func(context.Context, io.Writer, template.HTML) error
}

// Build compiles the ContainerHTML component by rendering children and executing the HTML layout callback.
func (e ContainerHTML) Build(cat Catalog, ctx context.Context, w io.Writer) error {
	children, err := RenderChildren(cat, ctx, e.Children)
	if err != nil {
		return err
	}
	if e.HTML != nil {
		return e.HTML(ctx, w, children)
	}
	_, err = io.WriteString(w, string(children))
	return err
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
