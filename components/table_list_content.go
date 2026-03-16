package components

import (
	"context"
	"fmt"

	"github.com/lariv-in/getters"
	. "maragu.dev/gomponents"
	g_html "maragu.dev/gomponents/html"
)

type TableListContent[T any] struct {
	Page
	Columns []TableColumn
	Data    getters.Getter[ObjectList[T]]
	OnClick getters.Getter[string]
}

func (e TableListContent[T]) Build(ctx context.Context) Node {
	var data ObjectList[T]
	if e.Data != nil {
		resolved, err := e.Data(ctx)
		if err == nil {
			data = resolved
		}
	}

	var ths []Node
	for _, col := range e.Columns {
		ths = append(ths, g_html.Th(g_html.Class("whitespace-nowrap min-w-[100px]"), Text(col.Label)))
	}

	var trs []Node
	if len(data.Items) == 0 {
		trs = append(trs, g_html.Tr(g_html.Td(g_html.ColSpan(fmt.Sprintf("%d", len(e.Columns))), g_html.Class("text-center opacity-50 py-8"), Text("Table is empty"))))
	} else {
		for _, row := range data.Items {
			rowMap := getters.MapFromStruct(row)
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
				expr, err := e.OnClick(rowCtx)
				if err == nil && expr != "" {
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

func (e TableListContent[T]) GetKey() string {
	return e.Key
}

func (e TableListContent[T]) GetRoles() []string {
	return e.Roles
}

func (e TableListContent[T]) GetChildren() []PageInterface {
	children := []PageInterface{}
	for _, col := range e.Columns {
		children = append(children, col.Children...)
	}
	return children
}

func (e *TableListContent[T]) SetChildren(children []PageInterface) {
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
