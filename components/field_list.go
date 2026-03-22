package components

import (
	"context"
	"log/slog"

	"reflect"

	"github.com/lariv-in/lago/getters"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

type FieldList struct {
	Page
	Getter   getters.Getter[any] // resolves to a slice
	Classes  string
	Children []PageInterface // template for each item
}

func (e FieldList) Build(ctx context.Context) Node {
	var listNodes Group

	if e.Getter != nil {
		rawData, err := e.Getter(ctx)
		if err != nil {
			slog.Error("FieldList getter failed", "error", err, "key", e.Key)
			return ContainerError{Error: getters.GetterStatic(err)}.Build(ctx)
		}
		if rawData != nil {
			value := reflect.ValueOf(rawData)
			if value.Type().CanSeq2() {
				for _, item := range value.Seq2() {
					itemCtx := context.WithValue(ctx, "$row", item)
					var childrenNodes Group
					for _, child := range e.Children {
						childrenNodes = append(childrenNodes, Render(child, itemCtx))
					}
					listNodes = append(listNodes, Div(Class("list-item"), childrenNodes))
				}
			}
		}
	}

	return Div(Class(e.Classes), listNodes)
}

func (e FieldList) GetKey() string {
	return e.Key
}

func (e FieldList) GetRoles() []string {
	return e.Roles
}

func (e FieldList) GetChildren() []PageInterface {
	return e.Children
}

func (e *FieldList) SetChildren(children []PageInterface) {
	e.Children = children
}
