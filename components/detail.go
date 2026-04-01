package components

import (
	"context"
	"log/slog"
	"reflect"

	"github.com/lariv-in/lago/getters"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

// Detail binds an object from context to its descendants under "$in".
// Child fields can then resolve their values via GetterKey("$in.FieldName").
type Detail[T any] struct {
	Page
	Getter   getters.Getter[T]
	Children []PageInterface
}

func (e Detail[T]) Build(ctx context.Context) Node {
	childCtx := ctx
	if e.Getter != nil {
		value, err := e.Getter(ctx)
		if err != nil {
			slog.Error("Detail getter failed", "error", err, "key", e.Key)
			return ContainerError{Error: getters.Static(err)}.Build(ctx)
		}
		if v := reflect.ValueOf(value); v.IsValid() && !v.IsZero() {
			objMap := getters.MapFromStruct(value)
			childCtx = context.WithValue(ctx, "$in", objMap)
		}
	}

	var childNodes []Node
	for _, child := range e.Children {
		childNodes = append(childNodes, Render(child, childCtx))
	}
	return Div(Group(childNodes))
}

func (e Detail[T]) GetKey() string {
	return e.Key
}

func (e Detail[T]) GetRoles() []string {
	return e.Roles
}

func (e Detail[T]) GetChildren() []PageInterface {
	return e.Children
}

func (e *Detail[T]) SetChildren(children []PageInterface) {
	e.Children = children
}
