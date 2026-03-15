package components

import (
	"context"
	"fmt"
	"reflect"

	"github.com/lariv-in/getters"
	. "maragu.dev/gomponents"
	g_html "maragu.dev/gomponents/html"
)

type TableGridContent struct {
	Page
	Columns []TableColumn
	Data    getters.Getter
	OnClick getters.Getter
}

func (e TableGridContent) Build(ctx context.Context) Node {
	var objects []any
	data := getters.IfOrGetter(e.Data, ctx, nil)

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

	var cards []Node
	if len(objects) == 0 {
		cards = append(cards, g_html.Div(g_html.Class("col-span-full text-center opacity-50 py-8"), Text("Table is empty")))
	} else {
		for _, row := range objects {
			rowMap := getters.MapFromStruct(row)
			rowCtx := context.WithValue(ctx, "$row", rowMap)

			var contentNodes []Node
			// First column is the title
			if len(e.Columns) > 0 {
				var firstColNodes []Node
				for _, child := range e.Columns[0].Children {
					firstColNodes = append(firstColNodes, Render(child, rowCtx))
				}
				contentNodes = append(contentNodes, g_html.Div(g_html.Class("font-semibold text-md truncate"), Group(firstColNodes)))
			}

			// Remaining columns as small labels
			for _, col := range e.Columns[1:] {
				var colNodes []Node
				for _, child := range col.Children {
					colNodes = append(colNodes, Render(child, rowCtx))
				}
				contentNodes = append(contentNodes, g_html.Div(g_html.Class("text-sm flex gap-2 truncate"),
					g_html.Span(g_html.Class("font-semibold text-primary"), Text(col.Label+":")),
					g_html.Span(Group(colNodes)),
				))
			}

			if e.OnClick != nil {
				expr := fmt.Sprintf("%v", getters.IfOrGetter(e.OnClick, rowCtx, ""))
				if expr != "" {
					cards = append(cards, g_html.Div(
						g_html.Class("border border-base-300 rounded-box flex flex-col bg-base-100 p-2 cursor-pointer hover:bg-base-200"),
						Attr("@click", expr),
						Group(contentNodes),
					))
					continue
				}
			}
			cards = append(cards, g_html.Div(
				g_html.Class("border border-base-300 rounded-box flex flex-col bg-base-100 p-2"),
				Group(contentNodes),
			))
		}
	}

	return g_html.Div(g_html.Class("flex flex-col gap-4, @container"),
		g_html.Div(g_html.Class("overflow-x-auto"),
			g_html.Div(g_html.Class("grid grid-cols-1 @md:grid-cols-2 @2xl:grid-cols-3 @3xl:grid-cols-4 gap-2"),
				Group(cards),
			),
		),
	)
}

func (e TableGridContent) GetKey() string {
	return e.Key
}

func (e TableGridContent) GetRoles() []string {
	return e.Roles
}

func (e TableGridContent) GetChildren() []PageInterface {
	children := []PageInterface{}
	for _, col := range e.Columns {
		children = append(children, col.Children...)
	}
	return children
}

func (e *TableGridContent) SetChildren(children []PageInterface) {
	offset := 0
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
