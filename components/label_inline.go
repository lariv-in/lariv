package components

import (
	"context"
	"html/template"
	"io"
)

// LabelInline represents a layout component grouping a bold label prefix and child elements horizontally.
// It places the label (Title followed by ":") and the rendered children in a flex container row.
//
// Use Cases:
//   - Building metadata display rows (e.g., "Status: [FieldText]", "Project: [FieldLink]").
//
// Example:
//
//	 &components.LabelInline{
//	     Title: "Deploy Status",
//	     Children: []components.PageInterface{
//	         &components.FieldText{Getter: getters.Static("Healthy")},
//	     },
//	 }
type LabelInline struct {
	// Page embeds common component properties like Key and Roles.
	Page
	// Title represents the header prefix text displayed before the colon.
	Title string
	// Children represents the slice of nested sub-components rendered in the row.
	Children []PageInterface
	// Classes represents additional CSS classes applied to the output HTML wrapper.
	// (Discouraged: Use layout containers or theme styling instead of custom styling overrides).
	Classes string
}

// Build compiles the LabelInline component into a Div wrapping a primary label Span and nested children.
func (e LabelInline) Build(cat Catalog, ctx context.Context, w io.Writer) error {
	children, err := RenderChildren(cat, ctx, e.Children)
	if err != nil {
		return err
	}
	return Execute(w, "label_inline", struct {
		Classes  string
		Title    string
		Children template.HTML
	}{Classes: e.Classes, Title: e.Title, Children: children})
}

// GetKey returns the unique key identifier for this LabelInline component.
func (e LabelInline) GetKey() string {
	return e.Key
}

// GetRoles returns the authorized roles required to view this LabelInline.
func (e LabelInline) GetRoles() []string {
	return e.Roles
}

// GetChildren returns the slice of nested sub-components.
func (e LabelInline) GetChildren() []PageInterface {
	return e.Children
}

// SetChildren replaces the slice of nested sub-components.
func (e *LabelInline) SetChildren(children []PageInterface) {
	e.Children = children
}
