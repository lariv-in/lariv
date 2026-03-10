package components

import (
	"context"
	"fmt"

	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

type TableColumn struct {
	Label     string
	Key       string
	Orderable bool
	Children  []PageInterface
}

type DataTable struct {
	UID      string
	Columns  []TableColumn
	Data     Getter
	Title    string
	Subtitle string
	Classes  string
	// Displays is a map of view name to display component
	// e.g. "List": TableListContent, "Grid": TableGridContent
	Displays        map[string]func([]TableColumn, Getter, Getter) PageInterface
	FilterComponent PageInterface
	CreateUrl       Getter
	OnClick         Getter // Per-row Alpine @click expression (use GetterNavigate, GetterSelect, or GetterMultiSelect)
}

func (e DataTable) Build(ctx context.Context) Node {
	if e.Displays == nil {
		e.Displays = map[string]func([]TableColumn, Getter, Getter) PageInterface{
			"List": func(cols []TableColumn, data Getter, onClick Getter) PageInterface {
				return TableListContent{Columns: cols, Data: data, OnClick: onClick}
			},
			"Grid": func(cols []TableColumn, data Getter, onClick Getter) PageInterface {
				return TableGridContent{Columns: cols, Data: data, OnClick: onClick}
			},
		}
	}

	displayNodes := Group{}
	for name, builder := range e.Displays {
		displayNodes = append(displayNodes, Div(
			Attr("x-show", fmt.Sprintf("view === '%s'", name)),
			builder(e.Columns, e.Data, e.OnClick).Build(ctx),
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
			Attr("@click.outside", "$el.removeAttribute('open')"),
			El("summary", Class("btn btn-square dropdown-toggle btn-primary btn-sm"),
				Icon{Name: "funnel"}.Build(ctx),
			),
			Div(Class("card w-64 my-1.5 card-body shadow dropdown-content border border-base-300 rounded-box z-2 bg-base-100"),
				e.FilterComponent.Build(ctx),
			),
		)
	}

	// Create button
	var createNode Node
	if e.CreateUrl != nil {
		createUrl := fmt.Sprintf("%s", IfOrGetter(e.CreateUrl, ctx, ""))
		if createUrl != "" {
			createNode = A(Href(createUrl), Class("btn btn-square btn-outline btn-sm"),
				Icon{Name: "plus"}.Build(ctx),
			)
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
			displayNodes,
			TablePagination{Data: e.Data}.Build(ctx),
		),
	)
}

func (e DataTable) GetChildren() []PageInterface {
	var children []PageInterface
	if e.FilterComponent != nil {
		children = append(children, e.FilterComponent)
	}
	for _, col := range e.Columns {
		children = append(children, col.Children...)
	}
	return children
}
