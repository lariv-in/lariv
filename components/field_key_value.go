package components

import (
	"context"
	"fmt"
	"reflect"

	"github.com/lariv-in/getters"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

type FieldKeyValue struct {
	Page
	Getter     getters.Getter
	KeyField   string
	ValueField string
	Classes    string
}

func (e FieldKeyValue) Build(ctx context.Context) Node {
	raw := getters.IfOrGetter(e.Getter, ctx, nil)
	if raw == nil {
		return Div()
	}

	rv := reflect.ValueOf(raw)
	if rv.Kind() != reflect.Slice {
		return Div()
	}

	var nodes []Node
	for i := 0; i < rv.Len(); i++ {
		item := rv.Index(i).Interface()
		m := getters.MapFromStruct(item)
		k := fmt.Sprintf("%v", m[e.KeyField])
		v := fmt.Sprintf("%v", m[e.ValueField])
		nodes = append(nodes,
			Div(Class("mb-4 pb-4 border-b border-base-300 last:border-b-0"),
				Div(Class("font-medium text-sm text-base-content/70 mb-1"), Text(k)),
				Div(Class("whitespace-pre-wrap"), Text(v)),
			),
		)
	}
	return Div(Class(e.Classes), Group(nodes))
}
