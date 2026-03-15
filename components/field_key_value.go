package components

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/lariv-in/getters"
	"gorm.io/datatypes"
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

	jsonData, isJson := raw.(datatypes.JSON)
	if !isJson {
		return Div()
	}

	value, err := jsonData.Value()
	if err != nil {
		fmt.Println(err)
		return Div()
	}

	var val []map[string]string

	err = json.Unmarshal([]byte(value.(string)), &val)
	if err != nil {
		fmt.Println(err)
		return Div()
	}

	var nodes []Node
	for _, r := range val {
		k := r[e.KeyField]
		v := r[e.ValueField]
		nodes = append(nodes,
			Div(Class("mb-4 pb-4 border-b border-base-300 last:border-b-0"),
				Div(Class("font-medium text-sm text-base-content/70 mb-1"), Text(k)),
				Div(Class("whitespace-pre-wrap"), Text(v)),
			),
		)
	}
	return Div(Class(e.Classes), Group(nodes))
}
