package components

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/lariv-in/lago/getters"
	. "maragu.dev/gomponents"
	g_html "maragu.dev/gomponents/html"
)

type TableListContent[T any] struct {
	Page
	Columns []TableColumn
	Data    getters.Getter[ObjectList[T]]
	RowAttr getters.Getter[Node]
}

func (e TableListContent[T]) Build(ctx context.Context) Node {
	var data ObjectList[T]
	if e.Data != nil {
		resolved, err := e.Data(ctx)
		if err == nil {
			data = resolved
		}
	}

	req, hasReq := ctx.Value("$request").(*http.Request)
	var currentSort string
	if hasReq {
		currentSort = req.URL.Query().Get("sort")
	}

	var ths []Node
	for _, col := range e.Columns {
		if col.Name == "" || !hasReq {
			ths = append(ths, g_html.Th(g_html.Class("whitespace-nowrap min-w-[100px]"), Text(col.Label)))
			continue
		}
		sortURL := columnSortURL(req, col.Name)
		ind := sortColumnIndicator(currentSort, col.Name)
		ths = append(ths, g_html.Th(g_html.Class("whitespace-nowrap min-w-[100px]"),
			g_html.A(
				g_html.Href(sortURL),
				g_html.Class("link link-hover link-neutral no-underline hover:underline cursor-pointer font-inherit text-inherit inline-flex items-center gap-1"),
				Text(col.Label+ind),
			),
		))
	}

	var trs []Node
	if len(data.Items) == 0 {
		trs = append(trs, g_html.Tr(g_html.Td(g_html.ColSpan(fmt.Sprintf("%d", len(e.Columns))), g_html.Class("text-center opacity-50 py-8"), Text("Table is empty"))))
	} else {
		n := len(data.Items)
		for i, row := range data.Items {
			rowMap := getters.MapFromStruct(row)
			rowCtx := context.WithValue(ctx, "$row", rowMap)
			rowCtx = context.WithValue(rowCtx, getters.ContextKeyTableDisplay, getters.TableDisplayList)
			// Per-row list position for cell components (int; 0-based).
			rowCtx = context.WithValue(rowCtx, "$rowIndex", i)
			rowCtx = context.WithValue(rowCtx, "$isFirstRow", i == 0)
			rowCtx = context.WithValue(rowCtx, "$isLastRow", i == n-1)

			var tds []Node
			for _, col := range e.Columns {
				var cellNodes []Node
				for _, child := range col.Children {
					cellNodes = append(cellNodes, Render(child, rowCtx))
				}
				tds = append(tds, g_html.Td(g_html.Class("whitespace-nowrap truncate max-w-xs min-w-[100px]"), Group(cellNodes)))
			}

			var rowNodes []Node
			if e.RowAttr != nil {
				extra, err := e.RowAttr(rowCtx)
				if err != nil {
					return ContainerError{Error: getters.Static(err)}.Build(ctx)
				}
				if extra != nil {
					rowNodes = append(rowNodes, extra)
				} else {
					rowNodes = append(rowNodes, g_html.Class("hover:bg-base-200 transition-colors"))
				}
			} else {
				rowNodes = append(rowNodes, g_html.Class("hover:bg-base-200 transition-colors"))
			}
			rowNodes = append(rowNodes, Group(tds))
			trs = append(trs, g_html.Tr(rowNodes...))
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
		end := min(offset+n, len(children))
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

// columnSortURL preserves the current query string, cycles sort for the column
// (ASC → DESC → cleared), and resets page to 1. A different column always starts at ASC.
func columnSortURL(req *http.Request, columnKey string) string {
	current := req.URL.Query().Get("sort")
	next := nextSortClause(current, columnKey)
	u := *req.URL
	q := u.Query()
	if next == "" {
		q.Del("sort")
	} else {
		q.Set("sort", next)
	}
	q.Set("page", "1")
	u.RawQuery = q.Encode()
	return u.String()
}

func nextSortClause(current, key string) string {
	current = strings.TrimSpace(current)
	if current == "" {
		return key + " ASC"
	}
	parts := strings.Fields(current)
	if len(parts) == 0 {
		return key + " ASC"
	}
	curCol := parts[0]
	curDir := "ASC"
	if len(parts) >= 2 {
		curDir = strings.ToUpper(parts[len(parts)-1])
	}
	if strings.EqualFold(curCol, key) {
		if curDir == "DESC" {
			return ""
		}
		return key + " DESC"
	}
	return key + " ASC"
}

func sortColumnIndicator(currentSort, columnKey string) string {
	currentSort = strings.TrimSpace(currentSort)
	if currentSort == "" {
		return ""
	}
	parts := strings.Fields(currentSort)
	if len(parts) < 1 || !strings.EqualFold(parts[0], columnKey) {
		return ""
	}
	if len(parts) >= 2 && strings.ToUpper(parts[len(parts)-1]) == "DESC" {
		return " \u25BC"
	}
	return " \u25B2"
}
