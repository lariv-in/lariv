package components

import (
	"context"
	"fmt"
	"net/http"
	"sort"
	"strings"

	"github.com/lariv-in/lago/getters"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

// TableWithColumns is implemented by [DataTable] for [ButtonToggleColumns].
// TableWithColumns represents a table interface containing query columns configured with [TableColumn].
type TableWithColumns interface {
	PageInterface
	// TableColumns returns the configuration list of table columns.
	TableColumns() []TableColumn
}

// ParseEnabledTableColumnsParam parses a comma-separated list of [TableColumn.Name] parameters into a boolean set.
func ParseEnabledTableColumnsParam(s string) map[string]bool {
	if s == "" {
		return map[string]bool{}
	}
	parts := strings.Split(s, ",")
	m := make(map[string]bool)
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			m[p] = true
		}
	}
	return m
}

// FormatEnabledTableColumnsQuery encodes enabled column names into a comma-separated query string in stable order.
func FormatEnabledTableColumnsQuery(enabled map[string]bool) string {
	if len(enabled) == 0 {
		return ""
	}
	names := make([]string, 0, len(enabled))
	for name, on := range enabled {
		if on && name != "" {
			names = append(names, name)
		}
	}
	sort.Strings(names)
	return strings.Join(names, ",")
}

// FilterTableColumnsByEnabledMap filters columns returning only those with empty Name (always visible) or those marked as enabled in the map.
func FilterTableColumnsByEnabledMap(cols []TableColumn, enabled map[string]bool) []TableColumn {
	out := make([]TableColumn, 0, len(cols))
	for _, col := range cols {
		if col.Name == "" || enabled[col.Name] {
			out = append(out, col)
		}
	}
	return out
}

// GetterEnabledColumnsFromContext yields a Getter function retrieving the enabled columns set from context values.
func GetterEnabledColumnsFromContext(key string) getters.Getter[map[string]bool] {
	return func(ctx context.Context) (map[string]bool, error) {
		v := ctx.Value(key)
		if v == nil {
			return nil, nil
		}
		m, ok := v.(map[string]bool)
		if !ok {
			return nil, fmt.Errorf("context value %q: expected map[string]bool, got %T", key, v)
		}
		return m, nil
	}
}

func namedColumnSet(cols []TableColumn) map[string]bool {
	m := make(map[string]bool)
	for _, c := range cols {
		if c.Name != "" {
			m[c.Name] = true
		}
	}
	return m
}

func allNamedEnabled(cols []TableColumn, enabled map[string]bool) bool {
	for _, c := range cols {
		if c.Name == "" {
			continue
		}
		if !enabled[c.Name] {
			return false
		}
	}
	return true
}

// tableToggleColumnsURL updates queryKey and resets page=1, preserving other query parameters.
func tableToggleColumnsURL(req *http.Request, queryKey, toggleName string, cols []TableColumn) string {
	q := req.URL.Query()
	enabled := make(map[string]bool)
	if _, ok := q[queryKey]; ok {
		enabled = ParseEnabledTableColumnsParam(q.Get(queryKey))
	} else {
		enabled = namedColumnSet(cols)
	}
	if toggleName != "" {
		if enabled[toggleName] {
			delete(enabled, toggleName)
		} else {
			enabled[toggleName] = true
		}
	}
	u := *req.URL
	nq := u.Query()
	if len(namedColumnSet(cols)) == 0 || allNamedEnabled(cols, enabled) {
		nq.Del(queryKey)
	} else {
		v := FormatEnabledTableColumnsQuery(enabled)
		if v == "" {
			nq.Set(queryKey, "")
		} else {
			nq.Set(queryKey, v)
		}
	}
	nq.Set("page", "1")
	u.RawQuery = nq.Encode()
	return u.String()
}

// ButtonToggleColumns represents a table toolbar dropdown selector presenting checkbox toggles for column visibility.
// It modifies the current query URL parameter (comma-separated column names list) to dynamically toggle visible fields.
//
// Use Cases:
//   - Letting administrators hide/display fields (e.g. creation dates, secondary parameters) on active data grids.
//
// Example:
//
//	&components.ButtonToggleColumns{
//	    Table:    getters.Static(myDataTable),
//	    QueryKey: "cols",
//	}
type ButtonToggleColumns struct {
	// Page embeds common component properties like Key and Roles.
	Page
	// Table is the dynamic function retrieving the TableWithColumns component to toggle.
	Table getters.Getter[TableWithColumns]
	// QueryKey represents the query parameter string key representing column flags (e.g. "cols").
	QueryKey string
}

// GetKey returns the unique key identifier for this ButtonToggleColumns component.
func (e ButtonToggleColumns) GetKey() string {
	return e.Key
}

// GetRoles returns the authorized roles required to view this ButtonToggleColumns.
func (e ButtonToggleColumns) GetRoles() []string {
	return e.Roles
}

// Build compiles the ButtonToggleColumns component into an HTML dropdown selector element.
func (e ButtonToggleColumns) Build(ctx context.Context) Node {
	if e.Table == nil || e.QueryKey == "" {
		return Group{}
	}
	tab, err := e.Table(ctx)
	if err != nil {
		return ContainerError{Error: getters.Static(err)}.Build(ctx)
	}
	if tab == nil {
		return Group{}
	}
	cols := tab.TableColumns()
	req, ok := ctx.Value("$request").(*http.Request)
	if !ok {
		return Group{}
	}

	var rows []Node
	for _, col := range cols {
		if col.Name == "" {
			continue
		}
		q := req.URL.Query()
		checked := true
		if _, has := q[e.QueryKey]; has {
			checked = ParseEnabledTableColumnsParam(q.Get(e.QueryKey))[col.Name]
		}
		href := tableToggleColumnsURL(req, e.QueryKey, col.Name, cols)
		mark := "\u2610 "
		if checked {
			mark = "\u2611 "
		}
		rows = append(rows, A(
			Href(href),
			Class("link link-hover flex items-center gap-2 px-2 py-1 rounded hover:bg-base-200"),
			Text(mark+col.Label),
		))
	}
	if len(rows) == 0 {
		return Group{}
	}

	return El(
		"details",
		Class("dropdown dropdown-end"),
		Attr("@click.outside", "if(!$event.target.closest('.fk-modal-container')){$el.removeAttribute('open')}"),
		El("summary", Class("btn btn-square dropdown-toggle btn-primary btn-sm"), Render(Icon{Name: "view-columns"}, ctx)),
		Div(
			Class(tableButtonFilterDefaultContentClasses),
			Div(Class("text-sm font-semibold mb-2"), Text("Columns")),
			Group(rows),
		),
	)
}
