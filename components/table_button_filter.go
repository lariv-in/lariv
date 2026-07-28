package components

import (
	"context"
	"html/template"
	"io"
)

// target styling constant defining fallback DaisyUI dropdown details panel classes.
const tableButtonFilterDefaultContentClasses = "card w-64 my-1.5 card-body shadow dropdown-content border border-base-300 rounded-box z-2 bg-base-100"

// TableButtonFilter represents a filter dropdown container panel for DataTable.Actions.
// It displays a funnel icon summary toggle that displays nested child options (typically filter forms) in an absolute details block when clicked.
//
// Use Cases:
//   - Bundling complex search filters, selection dropdown checklists, or search query parameters inside tables toolbars.
//
// Example:
//
//	 &components.TableButtonFilter{
//	     Child: &components.FormComponent[FilterOptions]{...},
//	 }
type TableButtonFilter struct {
	// Page embeds common component properties like Key and Roles.
	Page
	// Child represents the nested sub-component (typically a filter form) rendered inside the dropdown panel.
	Child PageInterface
	// ContentClasses represents additional CSS classes applied to the output HTML wrapper.
	// (Discouraged: Use layout containers or theme styling instead of custom styling overrides).
	ContentClasses string
}

// GetKey returns the unique key identifier for this TableButtonFilter component.
func (e TableButtonFilter) GetKey() string {
	return e.Key
}

// GetRoles returns the authorized roles required to view this TableButtonFilter.
func (e TableButtonFilter) GetRoles() []string {
	return e.Roles
}

// GetChildren returns the slice of nested sub-components.
func (e TableButtonFilter) GetChildren() []PageInterface {
	if e.Child != nil {
		return []PageInterface{e.Child}
	}
	return nil
}

// SetChildren replaces the slice of nested sub-components.
func (e *TableButtonFilter) SetChildren(children []PageInterface) {
	if len(children) > 0 {
		e.Child = children[0]
	} else {
		e.Child = nil
	}
}

// Build compiles the TableButtonFilter component into a details HTML block wrapping dropdown items.
func (e TableButtonFilter) Build(cat Catalog, ctx context.Context, w io.Writer) error {
	contentClass := e.ContentClasses
	if contentClass == "" {
		contentClass = tableButtonFilterDefaultContentClasses
	}
	var panel template.HTML
	if e.Child != nil {
		h, err := RenderHTML(e.Child, cat, ctx)
		if err != nil {
			return err
		}
		panel = h
	}
	icon, err := RenderHTML(Icon{Name: "funnel"}, cat, ctx)
	if err != nil {
		return err
	}
	return Execute(w, "table_button_filter", struct {
		Icon         template.HTML
		ContentClass string
		Panel        template.HTML
	}{Icon: icon, ContentClass: contentClass, Panel: panel})
}
