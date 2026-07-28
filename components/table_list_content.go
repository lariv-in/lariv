package components

import (
	"context"
	"html/template"
	"io"
	"net/http"
	"strings"

	"github.com/lariv-in/lariv/getters"
)

// TableListContent represents a sub-component layout rendering query lists in standard HTML tabular tables.
// It sets up table headers with clickable sort cycles (ASC -> DESC -> none) containing sorting indicator marks (▲/▼), and renders zebra table row lines.
//
// Use Cases:
//   - Displaying large data sets in standard table layouts suited for widescreen desktop dashboards.
//
// Example:
//
//	&components.TableListContent[User]{
//	    Columns: userCols,
//	    Data:    userDataGetter,
//	    RowAttr: getters.RowAttrNavigate(lariv.RoutePath("users.DetailRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("$row.ID"))})),
//	}
type TableListContent[T any] struct {
	// Page embeds common component properties like Key and Roles.
	Page
	// Columns represents the slice of TableColumn configurations.
	Columns []TableColumn
	// Data represents the dynamic Getter retrieving the paginated ObjectList payload.
	Data getters.Getter[ObjectList[T]]
	// RowAttr represents the dynamic getter returning TR HTML attributes.
	RowAttr getters.Getter[HTMLAttributes]
}

type tableListHeader struct {
	Label   string
	SortURL string
	PushURL string
}

type tableListRow struct {
	Attrs HTMLAttributes
	Cells []template.HTML
}

// Build compiles the TableListContent component into tabular lists featuring sorting headers and zebra tr wrappers.
func (e TableListContent[T]) Build(cat Catalog, ctx context.Context, w io.Writer) error {
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

	headers := make([]tableListHeader, 0, len(e.Columns))
	for _, col := range e.Columns {
		h := tableListHeader{Label: col.Label}
		if col.Name != "" && hasReq {
			sortURL := columnSortURL(req, col.Name)
			ind := sortColumnIndicator(currentSort, col.Name)
			pushURL := "true"
			if strings.Contains(req.URL.Path, "select") {
				pushURL = "false"
			}
			h.Label = col.Label + ind
			h.SortURL = sortURL
			h.PushURL = pushURL
		}
		headers = append(headers, h)
	}

	if len(data.Items) == 0 {
		return Execute(w, "table_list_content", struct {
			Headers []tableListHeader
			Empty   bool
			ColSpan int
			Rows    []tableListRow
		}{
			Headers: headers,
			Empty:   true,
			ColSpan: len(e.Columns),
		})
	}

	rows := make([]tableListRow, 0, len(data.Items))
	n := len(data.Items)
	for i, row := range data.Items {
		rowMap := getters.MapFromStruct(row)
		rowCtx := context.WithValue(ctx, "$row", rowMap)
		rowCtx = context.WithValue(rowCtx, getters.ContextKeyTableDisplay, getters.TableDisplayList)
		rowCtx = context.WithValue(rowCtx, "$rowIndex", i)
		rowCtx = context.WithValue(rowCtx, "$isFirstRow", i == 0)
		rowCtx = context.WithValue(rowCtx, "$isLastRow", i == n-1)

		cells := make([]template.HTML, 0, len(e.Columns))
		for _, col := range e.Columns {
			cell, err := RenderChildren(cat, rowCtx, col.Children)
			if err != nil {
				return err
			}
			cells = append(cells, cell)
		}

		attrs, err := ResolveAttrs(rowCtx, e.RowAttr)
		if err != nil {
			return ContainerError{Error: getters.Static(err)}.Build(cat, ctx, w)
		}
		if len(attrs) == 0 {
			attrs = HTMLAttributes{"class": "hover:bg-base-200 transition-colors"}
		}
		rows = append(rows, tableListRow{Attrs: attrs, Cells: cells})
	}

	return Execute(w, "table_list_content", struct {
		Headers []tableListHeader
		Empty   bool
		ColSpan int
		Rows    []tableListRow
	}{
		Headers: headers,
		Empty:   false,
		ColSpan: len(e.Columns),
		Rows:    rows,
	})
}

// GetKey returns the unique key identifier for this TableListContent.
func (e TableListContent[T]) GetKey() string {
	return e.Key
}

// GetRoles returns the authorized roles required to view this TableListContent.
func (e TableListContent[T]) GetRoles() []string {
	return e.Roles
}

// GetChildren returns the slice of nested sub-components.
func (e TableListContent[T]) GetChildren() []PageInterface {
	children := []PageInterface{}
	for _, col := range e.Columns {
		children = append(children, col.Children...)
	}
	return children
}

// SetChildren replaces the slice of nested sub-components.
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
