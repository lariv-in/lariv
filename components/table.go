package components

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"

	"github.com/lariv-in/lago/getters"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

type TableColumn struct {
	Label string
	// Name is the column identifier used in the list view ?sort= query (e.g. "Name ASC").
	// When non-empty, the list header cycles sort: ascending, descending, then clears ?sort=.
	Name string
	// Orderable is reserved; header sorting is enabled whenever Name is non-empty.
	Orderable bool
	Children  []PageInterface
}

type DataTable[T any] struct {
	Page
	UID      string
	Columns  []TableColumn
	Data     getters.Getter[ObjectList[T]]
	Title    string
	Subtitle string
	Classes  string
	// Displays is a map of view name to display component
	// e.g. "List": TableListContent, "Grid": TableGridContent
	Displays map[string]func([]TableColumn, getters.Getter[ObjectList[T]], getters.Getter[string]) PageInterface
	// DefaultView is the initial display mode; must match a key in Displays. Empty means "List".
	DefaultView string
	// Actions are rendered in the toolbar after the view switcher (e.g. &TableButtonFilter{Child: ...}, &TableButtonCreate{Link: ...}).
	Actions []PageInterface
	OnClick getters.Getter[string] // Per-row Alpine @click expression (use GetterNavigate, GetterSelect)
}

func renderTableToolbarAction(ctx context.Context, a PageInterface) Node {
	if a == nil {
		return nil
	}
	switch v := a.(type) {
	case *ButtonLink:
		if v != nil && v.Link != nil {
			if u, err := v.Link(ctx); err == nil && u != "" {
				return Render(v, ctx)
			}
		}
		return nil
	case *TableButtonCreate:
		if v != nil && v.Link != nil {
			if u, err := v.Link(ctx); err == nil && u != "" {
				return Render(v, ctx)
			}
		}
		return nil
	default:
		return Render(a, ctx)
	}
}

func (e DataTable[T]) Build(ctx context.Context) Node {
	if e.Displays == nil {
		e.Displays = map[string]func([]TableColumn, getters.Getter[ObjectList[T]], getters.Getter[string]) PageInterface{
			"List": func(cols []TableColumn, data getters.Getter[ObjectList[T]], onClick getters.Getter[string]) PageInterface {
				return TableListContent[T]{Columns: cols, Data: data, OnClick: onClick}
			},
			"Grid": func(cols []TableColumn, data getters.Getter[ObjectList[T]], onClick getters.Getter[string]) PageInterface {
				return TableGridContent[T]{Columns: cols, Data: data, OnClick: onClick}
			},
		}
	}

	displayNodes := Group{}
	for name, builder := range e.Displays {
		displayNodes = append(displayNodes, Div(
			Attr("x-show", fmt.Sprintf("view === '%s'", name)), Render(builder(e.Columns, e.Data, e.OnClick), ctx),
		))
	}

	// View Switcher (Simple Select)
	var options []Node
	for name := range e.Displays {
		options = append(options, Option(Value(name), Text(name)))
	}

	var actionBar Group
	for _, a := range e.Actions {
		if n := renderTableToolbarAction(ctx, a); n != nil {
			actionBar = append(actionBar, n)
		}
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
				Group(actionBar),
			),
		),
		Div(Class("relative my-2"),
			displayNodes, Render(TablePagination[T]{Data: e.Data}, ctx),
		),
	)
}

func (e DataTable[T]) GetKey() string {
	return e.Key
}

func (e DataTable[T]) GetRoles() []string {
	return e.Roles
}

func (e DataTable[T]) GetChildren() []PageInterface {
	children := make([]PageInterface, 0, len(e.Actions))
	children = append(children, e.Actions...)
	for _, col := range e.Columns {
		children = append(children, col.Children...)
	}
	return children
}

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
