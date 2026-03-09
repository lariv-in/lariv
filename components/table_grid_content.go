package components

import (
	"context"
	"reflect"

	. "maragu.dev/gomponents"
	g_html "maragu.dev/gomponents/html"
)

type TableGridContent struct {
	Columns []TableColumn
	Data    Getter
}

func (e TableGridContent) Build(ctx context.Context) Node {
	var objects []any
	data := IfOrGetter(e.Data, ctx, nil)

	if data != nil {
		v := reflect.ValueOf(data)
		if v.Kind() == reflect.Slice || v.Kind() == reflect.Array {
			for i := 0; i < v.Len(); i++ {
				objects = append(objects, v.Index(i).Interface())
			}
		}
	}

	var cards []Node
	if len(objects) == 0 {
		cards = append(cards, g_html.Div(g_html.Class("col-span-full text-center opacity-50 py-8"), Text("Table is empty")))
	} else {
		for _, row := range objects {
			rowMap := MapFromStruct(row)
			rowCtx := context.WithValue(ctx, "$row", rowMap)

			var contentNodes []Node
			// First column is the title, font-semibold text-md truncate
			if len(e.Columns) > 0 {
				var firstColNodes []Node
				for _, child := range e.Columns[0].Children {
					firstColNodes = append(firstColNodes, child.Build(rowCtx))
				}
				contentNodes = append(contentNodes, g_html.Div(g_html.Class("font-semibold text-md truncate"), Group(firstColNodes)))
			}

			// Remaining columns as small labels
			for _, col := range e.Columns[1:] {
				var colNodes []Node
				for _, child := range col.Children {
					colNodes = append(colNodes, child.Build(rowCtx))
				}
				contentNodes = append(contentNodes, g_html.Div(g_html.Class("text-sm flex gap-2 truncate"),
					g_html.Span(g_html.Class("font-semibold text-primary"), Text(col.Label+":")),
					g_html.Span(Group(colNodes)),
				))
			}

			cards = append(cards, g_html.Div(
				g_html.Class("border border-base-300 rounded-box flex flex-col bg-base-100 p-2"),
				Group(contentNodes),
			))
		}
	}

	return g_html.Div(g_html.Class("flex flex-col gap-4"),
		g_html.Div(g_html.Class("overflow-x-auto"),
			g_html.Div(g_html.Class("grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-2"),
				Group(cards),
			),
		),
	)
}

func (e TableGridContent) GetChildren() []PageInterface {
	children := []PageInterface{}
	for _, col := range e.Columns {
		children = append(children, col.Children...)
	}
	return children
}
