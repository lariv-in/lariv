package components

import (
	"context"
	"fmt"
	"reflect"

	. "maragu.dev/gomponents"
	g_html "maragu.dev/gomponents/html"
)

type TableListContent struct {
	Page
	Columns []TableColumn
	Data    Getter
	OnClick Getter
}

func (e TableListContent) Build(ctx context.Context) Node {
	var objects []any
	data := IfOrGetter(e.Data, ctx, nil)

	if data != nil {
		v := reflect.ValueOf(data)
		if v.Kind() == reflect.Pointer || v.Kind() == reflect.Interface {
			v = v.Elem()
		}

		if v.Kind() == reflect.Struct {
			itemsField := v.FieldByName("Items")
			if itemsField.IsValid() && (itemsField.Kind() == reflect.Slice || itemsField.Kind() == reflect.Array) {
				for i := 0; i < itemsField.Len(); i++ {
					objects = append(objects, itemsField.Index(i).Interface())
				}
			}
		}
	}

	var ths []Node
	for _, col := range e.Columns {
		ths = append(ths, g_html.Th(g_html.Class("whitespace-nowrap min-w-[100px]"), Text(col.Label)))
	}

	var trs []Node
	if len(objects) == 0 {
		trs = append(trs, g_html.Tr(g_html.Td(g_html.ColSpan(fmt.Sprintf("%d", len(e.Columns))), g_html.Class("text-center opacity-50 py-8"), Text("Table is empty"))))
	} else {
		for _, row := range objects {
			rowMap := MapFromStruct(row)
			rowCtx := context.WithValue(ctx, "$row", rowMap)

			var tds []Node
			for _, col := range e.Columns {
				var cellNodes []Node
				for _, child := range col.Children {
					cellNodes = append(cellNodes, Render(child, rowCtx))
				}
				tds = append(tds, g_html.Td(g_html.Class("whitespace-nowrap truncate max-w-xs min-w-[100px]"), Group(cellNodes)))
			}

			if e.OnClick != nil {
				expr := fmt.Sprintf("%v", IfOrGetter(e.OnClick, rowCtx, ""))
				if expr != "" {
					trs = append(trs, g_html.Tr(
						g_html.Class("cursor-pointer hover:bg-base-200 transition-colors"),
						Attr("@click", expr),
						Group(tds),
					))
					continue
				}
			}
			trs = append(trs, g_html.Tr(g_html.Class("hover:bg-base-200 transition-colors"), Group(tds)))
		}
	}

	return g_html.Div(g_html.Class("table-container flex flex-col rounded-box border border-base-300 bg-base-100"),
		g_html.Div(g_html.Class("overflow-x-auto"),
			g_html.Table(g_html.Class("table table-zebra"),
				g_html.THead(g_html.Tr(ths...)),
				g_html.TBody(trs...),
			),
		),
	)
}

func (e TableListContent) GetChildren() []PageInterface {
	children := []PageInterface{}
	for _, col := range e.Columns {
		children = append(children, col.Children...)
	}
	return children
}
