package components

import (
	"context"
	"html/template"
	"io"

	"github.com/lariv-in/lariv/getters"
)

// TableGridContent represents a sub-component layout that renders list data rows as a grid of responsive cards.
// It maps the first column value as the card header title and displays subsequent column fields as small labeled text blocks inside the card block.
//
// Use Cases:
//   - Rendering resource collections in card grid views on responsive layouts, suitable for products or user profile items.
//
// Example:
//
//	&components.TableGridContent[Product]{
//	    Columns: productCols,
//	    Data:    productDataGetter,
//	    RowAttr: getters.RowAttrNavigate(lariv.RoutePath("products.DetailRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("$row.ID"))})),
//	}
type TableGridContent[T any] struct {
	// Page embeds common component properties like Key and Roles.
	Page
	// Columns represents the slice of TableColumn configurations mapping data rows.
	Columns []TableColumn
	// Data represents the dynamic Getter retrieving the paginated ObjectList payload.
	Data getters.Getter[ObjectList[T]]
	// RowAttr represents the dynamic getter returning card HTML attributes.
	RowAttr getters.Getter[HTMLAttributes]
}

type tableGridField struct {
	Label   string
	Content template.HTML
}

type tableGridCard struct {
	Attrs  HTMLAttributes
	Title  template.HTML
	Fields []tableGridField
}

// Build compiles the TableGridContent component into a grid of card blocks.
func (e TableGridContent[T]) Build(cat Catalog, ctx context.Context, w io.Writer) error {
	var data ObjectList[T]
	if e.Data != nil {
		resolved, err := e.Data(ctx)
		if err == nil {
			data = resolved
		}
	}

	if len(data.Items) == 0 {
		return Execute(w, "table_grid_content", struct {
			Empty bool
			Cards []tableGridCard
		}{Empty: true})
	}

	cards := make([]tableGridCard, 0, len(data.Items))
	n := len(data.Items)
	for i, row := range data.Items {
		rowMap := getters.MapFromStruct(row)
		rowCtx := context.WithValue(ctx, "$row", rowMap)
		rowCtx = context.WithValue(rowCtx, getters.ContextKeyTableDisplay, getters.TableDisplayGrid)
		rowCtx = context.WithValue(rowCtx, "$rowIndex", i)
		rowCtx = context.WithValue(rowCtx, "$isFirstRow", i == 0)
		rowCtx = context.WithValue(rowCtx, "$isLastRow", i == n-1)

		var title template.HTML
		if len(e.Columns) > 0 {
			t, err := RenderChildren(cat, rowCtx, e.Columns[0].Children)
			if err != nil {
				return err
			}
			title = t
		}

		fields := make([]tableGridField, 0, max(len(e.Columns)-1, 0))
		for _, col := range e.Columns[1:] {
			content, err := RenderChildren(cat, rowCtx, col.Children)
			if err != nil {
				return err
			}
			fields = append(fields, tableGridField{Label: col.Label, Content: content})
		}

		attrs, err := ResolveAttrs(rowCtx, e.RowAttr)
		if err != nil {
			return ContainerError{Error: getters.Static(err)}.Build(cat, ctx, w)
		}
		if len(attrs) == 0 {
			attrs = HTMLAttributes{"class": "border border-base-300 rounded-box flex flex-col bg-base-100 p-2 hover:bg-base-200 transition-colors"}
		}
		cards = append(cards, tableGridCard{Attrs: attrs, Title: title, Fields: fields})
	}

	return Execute(w, "table_grid_content", struct {
		Empty bool
		Cards []tableGridCard
	}{Empty: false, Cards: cards})
}

// GetKey returns the unique key identifier for this TableGridContent.
func (e TableGridContent[T]) GetKey() string {
	return e.Key
}

// GetRoles returns the authorized roles required to view this TableGridContent.
func (e TableGridContent[T]) GetRoles() []string {
	return e.Roles
}

// GetChildren returns the slice of nested sub-components.
func (e TableGridContent[T]) GetChildren() []PageInterface {
	children := []PageInterface{}
	for _, col := range e.Columns {
		children = append(children, col.Children...)
	}
	return children
}

// SetChildren replaces the slice of nested sub-components.
func (e *TableGridContent[T]) SetChildren(children []PageInterface) {
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
