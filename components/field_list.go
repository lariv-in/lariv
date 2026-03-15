package components

import (
	"context"

	"github.com/lariv-in/getters"
	"reflect"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

type FieldList struct {
	Page
	Getter   getters.Getter  // resolves to a slice
	Classes  string
	Children []PageInterface // template for each item
}

func (e FieldList) Build(ctx context.Context) Node {
	var listNodes Group

	if e.Getter != nil {
		if rawData := getters.IfOrGetter(e.Getter, ctx, nil); rawData != nil {
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

func (e FieldList) GetChildren() []PageInterface {
	return e.Children
}
