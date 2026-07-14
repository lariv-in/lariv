package components

import (
	"context"

	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
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
	Child          PageInterface
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
func (e TableButtonFilter) Build(ctx context.Context) Node {
	contentClass := e.ContentClasses
	if contentClass == "" {
		contentClass = tableButtonFilterDefaultContentClasses
	}
	var panel Node = Group{}
	if e.Child != nil {
		panel = Render(e.Child, ctx)
	}
	return El("details",
		Class("dropdown dropdown-end"),
		Attr("@click.outside", "if(!$event.target.closest('.fk-modal-container')){$el.removeAttribute('open')}"),
		El("summary", Class("btn btn-square dropdown-toggle btn-primary btn-sm"), Render(Icon{Name: "funnel"}, ctx)),
		Div(Class(contentClass), panel),
	)
}
