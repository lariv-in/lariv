package components

import (
	"context"
	"html/template"
	"io"
)

// LabelNewline represents a layout component grouping a bold label prefix and child elements vertically.
// It places the label (Title followed by ":") and the rendered children in a vertical column container.
//
// Use Cases:
//   - Displaying larger content blocks or descriptions under a labeled header (e.g. "Description: [FieldTextArea/Paragraph]").
//
// Example:
//
//	 &components.LabelNewline{
//	     Title: "Biography",
//	     Children: []components.PageInterface{
//	         &components.FieldText{Getter: getters.Static("Full stack software engineer.")},
//	     },
//	 }
type LabelNewline struct {
	// Page embeds common component properties like Key and Roles.
	Page
	// Title represents the header label text displayed above the children.
	Title string
	// Children represents the slice of nested sub-components rendered below the title.
	Children []PageInterface
	// Classes represents additional CSS classes applied to the output HTML wrapper.
	// (Discouraged: Use layout containers or theme styling instead of custom styling overrides).
	Classes string
}

// Build compiles the LabelNewline component into a Div wrapping a bold Title label and children below it.
func (e LabelNewline) Build(cat Catalog, ctx context.Context, w io.Writer) error {
	children, err := RenderChildren(cat, ctx, e.Children)
	if err != nil {
		return err
	}
	return Execute(w, "label_newline", struct {
		Classes  string
		Title    string
		Children template.HTML
	}{Classes: e.Classes, Title: e.Title, Children: children})
}

// GetKey returns the unique key identifier for this LabelNewline component.
func (e LabelNewline) GetKey() string {
	return e.Key
}

// GetRoles returns the authorized roles required to view this LabelNewline.
func (e LabelNewline) GetRoles() []string {
	return e.Roles
}

// GetChildren returns the slice of nested sub-components.
func (e LabelNewline) GetChildren() []PageInterface {
	return e.Children
}

// SetChildren replaces the slice of nested sub-components.
func (e *LabelNewline) SetChildren(children []PageInterface) {
	e.Children = children
}
