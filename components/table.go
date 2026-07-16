package components

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"

	"github.com/lariv-in/lariv/getters"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

// TableColumn represents a single column header and cell structure layout configuration.
type TableColumn struct {
	// Label represents the header label text displayed at the top of the column.
	Label string
	// Name represents the query parameter sort value mapping to this column (e.g. "Name ASC").
	// Sorting headers are automatically enabled when this property is non-empty.
	Name string
	// Orderable is reserved for future sort config extensions.
	Orderable bool
	// Children represents the list of cell sub-components rendering inside this column's cells.
	Children []PageInterface
}

// TableDisplayBuilder defines the constructor callback type constructing list view mode components.
type TableDisplayBuilder[T any] func([]TableColumn, getters.Getter[ObjectList[T]], getters.Getter[Node]) PageInterface

// DataTable represents a responsive data grid list viewer component supporting paginated queries.
// It displays database rows in multiple view modes (e.g. List, Grid) using Alpine.js, handles sorting, pagination controls,
// dynamic column visibility filtering, and rendering toolbar action handlers.
//
// Use Cases:
//   - Showing paginated resource collections (users, audit logs, devices) featuring query sorting and grid/list toggles.
//
// Example:
//
//	&components.DataTable[User]{
//	    Title:    "Accounts List",
//	    Subtitle: "Manage system user credentials",
//	    Columns: []components.TableColumn{
//	        {Label: "Email Address", Name: "email"},
//	        {Label: "Roles", Name: "roles"},
//	    },
//	    Data: userDataGetter,
//	}
type DataTable[T any] struct {
	// Page embeds common component properties like Key and Roles.
	Page
	// UID represents the unique HTML element wrapper ID (defaults to "table-container").
	UID string
	// Columns represents the slice of TableColumn configurations.
	Columns []TableColumn
	// Data represents the dynamic Getter retrieving the paginated ObjectList payload.
	Data getters.Getter[ObjectList[T]]
	// Title represents the heading label text displayed above the table panel.
	Title string
	// Subtitle represents the auxiliary description label text displayed below the title.
	Subtitle string
	// Classes represents additional CSS classes applied to the output HTML wrapper.
	// (Discouraged: Use layout containers or theme styling instead of custom styling overrides).
	Classes string
	// Displays maps view mode names (e.g. "List", "Grid") to their constructor callback builders.
	Displays map[string]TableDisplayBuilder[T]
	// DefaultView represents the initial active view mode key (defaults to "List").
	DefaultView string
	// Actions represents custom toolbar action buttons rendered next to view switch controls.
	Actions []PageInterface
	// RowAttr represents the dynamic getter returning TR/card attribute nodes.
	RowAttr getters.Getter[Node]
	// EnabledColumns represents the dynamic getter returning the visible columns map filter.
	EnabledColumns getters.Getter[map[string]bool]
}

// TableColumns returns the configuration list of table columns.
func (e DataTable[T]) TableColumns() []TableColumn {
	return e.Columns
}

// Build compiles the DataTable component into table headers, lists, grids, and view switchers.
func (e DataTable[T]) Build(ctx context.Context) Node {
	if e.Displays == nil {
		e.Displays = map[string]TableDisplayBuilder[T]{
			"List": func(cols []TableColumn, data getters.Getter[ObjectList[T]], rowAttr getters.Getter[Node]) PageInterface {
				return TableListContent[T]{Columns: cols, Data: data, RowAttr: rowAttr}
			},
			"Grid": func(cols []TableColumn, data getters.Getter[ObjectList[T]], rowAttr getters.Getter[Node]) PageInterface {
				return TableGridContent[T]{Columns: cols, Data: data, RowAttr: rowAttr}
			},
		}
	}

	displayCols := e.Columns
	if e.EnabledColumns != nil {
		enabledMap, err := e.EnabledColumns(ctx)
		if err != nil {
			return ContainerError{Error: getters.Static(err)}.Build(ctx)
		}
		if enabledMap != nil {
			displayCols = FilterTableColumnsByEnabledMap(e.Columns, enabledMap)
		}
	}

	displayNodes := Group{}
	for name, builder := range e.Displays {
		displayNodes = append(displayNodes, Div(
			Attr("x-show", fmt.Sprintf("view === '%s'", name)), Render(builder(displayCols, e.Data, e.RowAttr), ctx),
		))
	}

	// View Switcher (Simple Select)
	var options []Node
	for name := range e.Displays {
		options = append(options, Option(Value(name), Text(name)))
	}

	var actionBar Group
	for _, a := range e.Actions {
		if a == nil {
			continue
		}
		actionBar = append(actionBar, Render(a, ctx))
	}

	uid := e.UID
	if uid == "" {
		uid = "table-container"
	}

	initialView := e.DefaultView
	if initialView == "" {
		initialView = "List"
	}
	if _, ok := e.Displays[initialView]; !ok {
		if _, ok := e.Displays["List"]; ok {
			initialView = "List"
		} else {
			keys := make([]string, 0, len(e.Displays))
			for k := range e.Displays {
				keys = append(keys, k)
			}
			sort.Strings(keys)
			if len(keys) > 0 {
				initialView = keys[0]
			}
		}
	}
	xData, _ := json.Marshal(map[string]string{"view": initialView})

	return Div(
		ID(uid), Class(fmt.Sprintf("w-full data-table-container %s", e.Classes)),
		Attr("x-data", string(xData)),
		Div(
			Class("flex justify-between items-center my-2"),
			Div(
				If(e.Title != "", Div(Class("text-xl font-semibold"), Text(e.Title))),
				If(e.Subtitle != "", Div(Class("text-sm text-gray-500"), Text(e.Subtitle))),
			),
			Div(
				Class("flex items-center gap-2"),
				Select(
					Class("select select-md"),
					Attr("x-model", "view"),
					Group(options),
				),
				Group(actionBar),
			),
		),
		Div(
			Class("relative my-2"),
			displayNodes, Render(TablePagination[T]{Data: e.Data}, ctx),
		),
	)
}

// GetKey returns the unique key identifier for this DataTable.
func (e DataTable[T]) GetKey() string {
	return e.Key
}

// GetRoles returns the authorized roles required to view this DataTable.
func (e DataTable[T]) GetRoles() []string {
	return e.Roles
}

// GetChildren returns the slice of nested sub-components.
func (e DataTable[T]) GetChildren() []PageInterface {
	children := make([]PageInterface, 0, len(e.Actions))
	children = append(children, e.Actions...)
	for _, col := range e.Columns {
		children = append(children, col.Children...)
	}
	return children
}

// SetChildren replaces the slice of nested sub-components.
func (e *DataTable[T]) SetChildren(children []PageInterface) {
	offset := 0
	for i := range e.Actions {
		if offset >= len(children) {
			return
		}
		e.Actions[i] = children[offset]
		offset++
	}
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
