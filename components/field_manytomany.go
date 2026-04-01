package components

import (
	"context"
	"log/slog"

	"github.com/lariv-in/lago/getters"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

// FieldManyToMany renders a read-only list of related records for detail views.
// Use the same Getter and Display getters as InputManyToMany[T] on the matching form.
//
// If Link is set, it is resolved with getters.ContextKeyIn bound to each related
// record (same as Display), e.g. lago.GetterRoutePath(..., {"id": GetterKey[uint]("$in.ID")}).
type FieldManyToMany[T any] struct {
	Page
	Label     string
	Getter    getters.Getter[[]T]
	Display   getters.Getter[string]
	Link      getters.Getter[string]
	Classes   string
	EmptyText string
}

func (e FieldManyToMany[T]) GetKey() string {
	return e.Key
}

func (e FieldManyToMany[T]) GetRoles() []string {
	return e.Roles
}

func (e FieldManyToMany[T]) Build(ctx context.Context) Node {
	var chipNodes []Node
	if e.Getter != nil {
		values, err := e.Getter(ctx)
		if err != nil {
			slog.Error("FieldManyToMany getter failed", "error", err, "key", e.Key)
			return ContainerError{Error: getters.Static(err)}.Build(ctx)
		}
		for _, v := range values {
			pair, ok := manyToManySelectionPair(ctx, v, e.Display, e.Key)
			if !ok {
				continue
			}
			label := Span(Class("text-sm flex-1 min-w-0 truncate"), Text(pair.Value))
			chipClass := "flex items-center gap-1 rounded-lg bg-base-200 pl-2 pr-2 py-1 min-w-0 max-w-full"
			if e.Link != nil {
				itemCtx := context.WithValue(ctx, getters.ContextKeyIn, getters.MapFromStruct(v))
				href, err := e.Link(itemCtx)
				if err != nil {
					slog.Error("FieldManyToMany link getter failed", "error", err, "key", e.Key)
				} else if href != "" {
					chipNodes = append(chipNodes, A(
						Href(href),
						Class(chipClass+" link link-hover no-underline"),
						label,
					))
					continue
				}
			}
			chipNodes = append(chipNodes, Div(Class(chipClass), label))
		}
	}

	empty := e.EmptyText
	if empty == "" {
		empty = "—"
	}

	var body Node
	if len(chipNodes) == 0 {
		body = Span(Class("text-sm opacity-50"), Text(empty))
	} else {
		body = Div(
			Class("input input-bordered w-full min-h-12 h-auto flex flex-wrap items-center gap-2"),
			Group(chipNodes),
		)
	}

	outer := []Node{body}
	if e.Label != "" {
		outer = append([]Node{Label(Class("label text-sm font-bold"), Text(e.Label))}, outer...)
	}

	return Div(Class("my-1 "+e.Classes), Group(outer))
}
