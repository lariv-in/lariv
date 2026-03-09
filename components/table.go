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
	Columns  []TableColumn
	Data     Getter
	Title    string
	Subtitle string
	// Displays is a map of view name to display component
	// e.g. "List": TableListContent, "Grid": TableGridContent
	Displays map[string]func([]TableColumn, Getter) PageInterface
}

func (e DataTable) Build(ctx context.Context) Node {
	if e.Displays == nil {
		e.Displays = map[string]func([]TableColumn, Getter) PageInterface{
			"List": func(cols []TableColumn, data Getter) PageInterface {
				return TableListContent{Columns: cols, Data: data}
			},
			"Grid": func(cols []TableColumn, data Getter) PageInterface {
				return TableGridContent{Columns: cols, Data: data}
			},
		}
	}

	displayNodes := Group{}
	for name, builder := range e.Displays {
		displayNodes = append(displayNodes, Div(
			Attr("x-show", fmt.Sprintf("view === '%s'", name)),
			builder(e.Columns, e.Data).Build(ctx),
		))
	}

	// View Switcher (Simple Select)
	var options []Node
	for name := range e.Displays {
		options = append(options, Option(Value(name), Text(name)))
	}

	return Div(
		ID("table-container"), Class("w-full"),
		Attr("x-data", "{ view: 'List' }"),
		Div(Class("flex justify-between items-center mb-4"),
			Div(
				H2(Class("text-xl font-bold"), Text(e.Title)),
				P(Class("text-sm text-gray-500"), Text(e.Subtitle)),
			),
			Div(Class("flex items-center gap-2"),
				Select(Class("select select-sm select-bordered"),
					Attr("x-model", "view"),
					Group(options),
				),
			),
		),
		displayNodes,
	)
}

func (e DataTable) GetChildren() []PageInterface {
	return nil
}
