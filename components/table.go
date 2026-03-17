package components

import (
	"context"
	"fmt"

	"github.com/lariv-in/getters"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

type TableColumn struct {
	Label     string
	Key       string
	Orderable bool
	Children  []PageInterface
}

type DataTable[T any] struct {
	Page
	UID      string
	Columns  []TableColumn
	Data     getters.Getter[ObjectList[T]]
	Title    string
	Subtitle string
	Classes  string
	// Displays is a map of view name to display component
	// e.g. "List": TableListContent, "Grid": TableGridContent
	Displays        map[string]func([]TableColumn, getters.Getter[ObjectList[T]], getters.Getter[string]) PageInterface
	FilterComponent PageInterface
	CreateUrl       getters.Getter[string]
	OnClick         getters.Getter[string] // Per-row Alpine @click expression (use GetterNavigate, GetterSelect, or GetterMultiSelect)
}

func (e DataTable[T]) Build(ctx context.Context) Node {
	if e.Displays == nil {
		e.Displays = map[string]func([]TableColumn, getters.Getter[ObjectList[T]], getters.Getter[string]) PageInterface{
			"List": func(cols []TableColumn, data getters.Getter[ObjectList[T]], onClick getters.Getter[string]) PageInterface {
				return TableListContent[T]{Columns: cols, Data: data, OnClick: onClick}
			},
			"Grid": func(cols []TableColumn, data getters.Getter[ObjectList[T]], onClick getters.Getter[string]) PageInterface {
				return TableGridContent[T]{Columns: cols, Data: data, OnClick: onClick}
			},
		}
	}

	displayNodes := Group{}
	for name, builder := range e.Displays {
		displayNodes = append(displayNodes, Div(
			Attr("x-show", fmt.Sprintf("view === '%s'", name)), Render(builder(e.Columns, e.Data, e.OnClick), ctx),
		))
	}

	// View Switcher (Simple Select)
	var options []Node
	for name := range e.Displays {
		options = append(options, Option(Value(name), Text(name)))
	}

	// Filter dropdown
	var filterNode Node
	if e.FilterComponent != nil {
		filterNode = El("details",
			Class("dropdown dropdown-end"),
			Attr("@click.outside", "if(!$event.target.closest('.fk-modal-container')){$el.removeAttribute('open')}"),
			El("summary", Class("btn btn-square dropdown-toggle btn-primary btn-sm"), Render(Icon{Name: "funnel"}, ctx)),
			Div(Class("card w-64 my-1.5 card-body shadow dropdown-content border border-base-300 rounded-box z-2 bg-base-100"), Render(e.FilterComponent, ctx)),
		)
	}

	// Create button
	var createNode Node
	if e.CreateUrl != nil {
		createURL, err := e.CreateUrl(ctx)
		if err == nil && createURL != "" {
			createNode = A(Href(createURL), Class("btn btn-square btn-outline btn-sm"), Render(Icon{Name: "plus"}, ctx))
		}
	}

	uid := e.UID
	if uid == "" {
		uid = "table-container"
	}

	return Div(
		ID(uid), Class(fmt.Sprintf("w-full data-table-container %s", e.Classes)),
		Attr("x-data", "{ view: 'List' }"),
		Div(Class("flex justify-between items-center my-2"),
			Div(
				If(e.Title != "", Div(Class("text-xl font-semibold"), Text(e.Title))),
				If(e.Subtitle != "", Div(Class("text-sm text-gray-500"), Text(e.Subtitle))),
			),
			Div(Class("flex items-center gap-2"),
				Select(Class("select select-md"),
					Attr("x-model", "view"),
					Group(options),
				),
				If(filterNode != nil, filterNode),
				If(createNode != nil, createNode),
			),
		),
		Div(Class("relative my-2"),
			displayNodes, Render(TablePagination[T]{Data: e.Data}, ctx),
		),
	)
}

func (e DataTable[T]) GetKey() string {
	return e.Key
}

func (e DataTable[T]) GetRoles() []string {
	return e.Roles
}

func (e DataTable[T]) GetChildren() []PageInterface {
	var children []PageInterface
	if e.FilterComponent != nil {
		children = append(children, e.FilterComponent)
	}
	for _, col := range e.Columns {
		children = append(children, col.Children...)
	}
	return children
}

func (e *DataTable[T]) SetChildren(children []PageInterface) {
	offset := 0
	if e.FilterComponent != nil && len(children) > 0 {
		e.FilterComponent = children[0]
		offset = 1
	}
	for i := range e.Columns {
		n := len(e.Columns[i].Children)
		end := offset + n
		if end > len(children) {
			end = len(children)
		}
		e.Columns[i].Children = children[offset:end]
		offset = end
		if offset >= len(children) {
			return
		}
	}
	if offset < len(children) && len(e.Columns) > 0 {
		e.Columns[len(e.Columns)-1].Children = append(e.Columns[len(e.Columns)-1].Children, children[offset:]...)
	}
}
