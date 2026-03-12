package components

import (
	"context"
	"fmt"
	"reflect"

	"github.com/lariv-in/getters"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

type Timeline struct {
	Page
	UID             string
	Title           string
	Classes         string
	Data            getters.Getter  // list of items
	OnClick         getters.Getter  // per-item URL (GetterNavigate)
	FilterComponent PageInterface   // optional filter form
	Children        []PageInterface // card content template
}

func (e Timeline) Build(ctx context.Context) Node {
	var data []any
	if e.Data != nil {
		if rawData := getters.IfOrGetter(e.Data, ctx, nil); rawData != nil {
			v := reflect.ValueOf(rawData)
			if v.Kind() == reflect.Pointer || v.Kind() == reflect.Interface {
				v = v.Elem()
			}
			if v.Kind() == reflect.Struct {
				itemsField := v.FieldByName("Items")
				if itemsField.IsValid() && (itemsField.Kind() == reflect.Slice || itemsField.Kind() == reflect.Array) {
					for i := 0; i < itemsField.Len(); i++ {
						data = append(data, itemsField.Index(i).Interface())
					}
				}
			}
		}
	}

	uid := e.UID
	if uid == "" {
		uid = "timeline-container"
	}

	var headerNode Node
	if e.Title != "" || e.FilterComponent != nil {
		var titleNode Node
		if e.Title != "" {
			titleNode = Div(Class("text-xl font-semibold"), Text(e.Title))
		}

		var filterNode Node
		if e.FilterComponent != nil {
			filterNode = El("details",
				Class("dropdown dropdown-end"),
				Attr("@click.outside", "$el.removeAttribute('open')"),
				El("summary", Class("btn btn-square dropdown-toggle btn-primary btn-sm"), Render(Icon{Name: "funnel"}, ctx)),
				Div(Class("card w-64 my-1.5 card-body shadow dropdown-content border border-base-300 rounded-box z-2 bg-base-100"), Render(e.FilterComponent, ctx)),
			)
		}

		headerNode = Div(Class("flex justify-between items-center mb-4"),
			titleNode,
			filterNode,
		)
	}

	var cardsGroup Group
	if len(data) == 0 {
		cardsGroup = append(cardsGroup, Div(Class("text-center opacity-60 py-8"), Text("No items found")))
	} else {
		for _, item := range data {
			rowMap := getters.MapFromStruct(item)
			itemCtx := context.WithValue(ctx, "$row", rowMap)

			var childrenNodes Group
			for _, child := range e.Children {
				childrenNodes = append(childrenNodes, Render(child, itemCtx))
			}

			var clickableClasses string

			timelineContent := Div(Class("timeline-item relative flex items-center gap-4 pb-6 last:pb-0"),
				Div(Class("timeline-indicator relative z-10 flex items-center"),
					Div(Class("w-3 h-3 rounded-full bg-primary")),
					Div(Class("h-0.5 w-4 bg-primary")),
				),
				Div(Class(fmt.Sprintf("timeline-card flex-1 p-4 rounded-box border border-base-300 bg-base-100 shadow-sm %s", clickableClasses)),
					childrenNodes,
				),
			)
			if e.OnClick != nil {
				url := fmt.Sprintf("%v", getters.IfOrGetter(e.OnClick, itemCtx, ""))
				if url != "" {
					timelineContent = A(Href(url), timelineContent)
				}
			}

			cardsGroup = append(cardsGroup,
				timelineContent,
			)
		}
	}

	var verticalLine Node
	if len(data) > 0 {
		verticalLine = Div(Class("absolute left-[5px] top-0 bottom-0 w-0.5 bg-primary opacity-30"))
	}

	return Div(ID(uid), Class(fmt.Sprintf("timeline-container %s", e.Classes)),
		headerNode,
		Div(Class("timeline-scroll relative"),
			verticalLine,
			cardsGroup,
		),
		Render(TablePagination{Data: e.Data}, ctx),
	)
}

func (e Timeline) GetChildren() []PageInterface {
	var children []PageInterface
	if e.FilterComponent != nil {
		children = append(children, e.FilterComponent)
	}
	children = append(children, e.Children...)
	return children
}
