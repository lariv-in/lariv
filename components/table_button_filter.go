package components

import (
	"context"

	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

const tableButtonFilterDefaultContentClasses = "card w-64 my-1.5 card-body shadow dropdown-content border border-base-300 rounded-box z-2 bg-base-100"

// TableButtonFilter is the default filter dropdown for DataTable.Actions (funnel summary + card panel).
// Set Child to the filter form or other content. ContentClasses overrides the inner panel classes when non-empty.
type TableButtonFilter struct {
	Page
	Child          PageInterface
	ContentClasses string
}

func (e TableButtonFilter) GetKey() string {
	return e.Key
}

func (e TableButtonFilter) GetRoles() []string {
	return e.Roles
}

func (e TableButtonFilter) GetChildren() []PageInterface {
	if e.Child != nil {
		return []PageInterface{e.Child}
	}
	return nil
}

func (e *TableButtonFilter) SetChildren(children []PageInterface) {
	if len(children) > 0 {
		e.Child = children[0]
	} else {
		e.Child = nil
	}
}

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
