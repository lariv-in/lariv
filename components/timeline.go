package components

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/lariv-in/lago/getters"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

type Timeline[T any] struct {
	Page
	UID             string
	Title           string
	Classes         string
	Data            getters.Getter[ObjectList[T]] // list of items
	OnClick         getters.Getter[string]        // per-item URL (GetterNavigate)
	FilterComponent PageInterface                 // optional filter form
	CreateUrl       getters.Getter[string]
	Children        []PageInterface // card content template
}

func (e Timeline[T]) Build(ctx context.Context) Node {
	var data []T
	if e.Data != nil {
		list, err := e.Data(ctx)
		if err != nil {
			slog.Error("Timeline Data getter failed", "error", err, "key", e.Key)
			return ContainerError{Error: getters.Static(err)}.Build(ctx)
		}
		data = list.Items
	}

	uid := e.UID
	if uid == "" {
		uid = "timeline-container"
	}

	var createNode Node
	if e.CreateUrl != nil {
		createURL, err := e.CreateUrl(ctx)
		if err == nil && createURL != "" {
			createNode = Render(ButtonLink{
				Link:    getters.Static(createURL),
				Icon:    "plus",
				Classes: "btn-square btn-outline btn-sm",
			}, ctx)
		}
	}

	var headerNode Node
	if e.Title != "" || e.FilterComponent != nil || createNode != nil {
		var filterNode Node
		if e.FilterComponent != nil {
			filterNode = El("details",
				Class("dropdown dropdown-end"),
				Attr("@click.outside", "$el.removeAttribute('open')"),
				El("summary", Class("btn btn-square dropdown-toggle btn-primary btn-sm"), Render(Icon{Name: "funnel"}, ctx)),
				Div(Class("card w-64 my-1.5 card-body shadow dropdown-content border border-base-300 rounded-box z-2 bg-base-100"), Render(e.FilterComponent, ctx)),
			)
		}

		var actions Group
		if filterNode != nil {
			actions = append(actions, filterNode)
		}
		if createNode != nil {
			actions = append(actions, createNode)
		}
		var actionsRow Node
		if len(actions) > 0 {
			actionsRow = Div(Class("flex items-center gap-2"), actions)
		}

		headerNode = Div(Class("flex justify-between items-center mb-4"),
			If(e.Title != "", Div(Class("text-xl font-semibold"), Text(e.Title))),
			actionsRow,
		)
	}

	var cardsGroup Group
	if len(data) == 0 {
		cardsGroup = append(cardsGroup, Div(Class("text-center opacity-60 py-8"), Text("No items found")))
	} else {
		for _, item := range data {
			rowMap := getters.MapFromStruct(any(item))
			itemCtx := context.WithValue(ctx, "$row", rowMap)

			var childrenNodes Group
			for _, child := range e.Children {
				childrenNodes = append(childrenNodes, Render(child, itemCtx))
			}

			var clickableClasses string

			timelineContent := Div(Class("timeline-item relative flex items-center gap-4"),
				Div(Class("timeline-indicator relative z-10 flex items-center"),
					Div(Class("w-3 h-3 rounded-full bg-primary")),
				),
				Div(Class(fmt.Sprintf("timeline-card flex-1 p-2 m-1 rounded-box border border-base-300 %s", clickableClasses)),
					childrenNodes,
				),
			)
			if e.OnClick != nil {
				url, err := e.OnClick(itemCtx)
				if err != nil {
					slog.Error("Timeline OnClick getter failed", "error", err, "key", e.Key)
					return ContainerError{Error: getters.Static(err)}.Build(ctx)
				}
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
		Render(TablePagination[T]{Data: e.Data}, ctx),
	)
}

func (e Timeline[T]) GetKey() string {
	return e.Key
}

func (e Timeline[T]) GetRoles() []string {
	return e.Roles
}

func (e Timeline[T]) GetChildren() []PageInterface {
	var children []PageInterface
	if e.FilterComponent != nil {
		children = append(children, e.FilterComponent)
	}
	children = append(children, e.Children...)
	return children
}

func (e *Timeline[T]) SetChildren(children []PageInterface) {
	offset := 0
	if e.FilterComponent != nil && len(children) > 0 {
		e.FilterComponent = children[0]
		offset = 1
	}
	e.Children = children[offset:]
}
