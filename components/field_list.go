package components

import (
	"context"
	"log/slog"

	"github.com/lariv-in/lago/getters"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

type FieldList[T any] struct {
	Page
	Getter   getters.Getter[[]T] // resolves to a slice
	Classes  string
	Children []PageInterface // template for each item
}

func (e FieldList[T]) Build(ctx context.Context) Node {
	var listNodes Group

	if e.Getter != nil {
		rawData, err := e.Getter(ctx)
		if err != nil {
			slog.Error("FieldList getter failed", "error", err, "key", e.Key)
			return ContainerError{Error: getters.Static(err)}.Build(ctx)
		}
		for _, item := range rawData {
			itemCtx := context.WithValue(ctx, "$row", item)
			var childrenNodes Group
			for _, child := range e.Children {
				childrenNodes = append(childrenNodes, Render(child, itemCtx))
			}
			listNodes = append(listNodes, Div(Class("list-item ml-4"), childrenNodes))
		}
	}

	return Div(Class(e.Classes), listNodes)
}

func (e FieldList[T]) GetKey() string {
	return e.Key
}

func (e FieldList[T]) GetRoles() []string {
	return e.Roles
}

func (e FieldList[T]) GetChildren() []PageInterface {
	return e.Children
}

func (e *FieldList[T]) SetChildren(children []PageInterface) {
	e.Children = children
}
