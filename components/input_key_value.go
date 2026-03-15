package components

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/lariv-in/getters"
	"github.com/lariv-in/registry"
	"gorm.io/datatypes"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

type InputKeyValue struct {
	Page
	Getter  getters.Getter
	Keys    getters.Getter
	Classes string
	Name    string
}

func (e InputKeyValue) Build(ctx context.Context) Node {
	raw := getters.IfOrGetter(e.Getter, ctx, nil)
	var val []registry.Pair[string, string]
	if raw != nil {
	value, _ := raw.(string)

	jsonData, isJson := raw.(datatypes.JSON)
	if isJson {
		jsonValue, err := jsonData.Value()
	if err != nil {
		fmt.Println(err)
		return Div()
	}
	value = jsonValue.(string)

	}


	err := json.Unmarshal([]byte(value), &val)
	if err != nil {
		fmt.Println(err)
		return Div()
	}
	}


	keys := e.Keys(ctx).([]string)

	var nodes []Node
	for i, k := range keys {
		isCurrentValue := true
		if i >= len(val) {
			isCurrentValue = false
		} else if val[i].Key != k {
			isCurrentValue = false
		}
		nodes = append(nodes, 
		InputText{Hidden: true, Name: e.Name + "Key", Getter: getters.GetterStatic(k)}.Build(ctx),
		InputTextarea{Name: e.Name + "Value",Label: k, Getter: func (ctx context.Context) any {
			if isCurrentValue {
				return val[i].Value
			}
			return nil
		}}.Build(ctx),
	)
	}

	finalInput := Input(
	Type("hidden"),
	Name(e.Name),
	Attr("x-data"),
	Attr("x-init", fmt.Sprintf(`
	$el.closest('form').addEventListener('submit', (e) => {
		let form = e.target;
		let data = [];
		let fd = new FormData(form);
            let keys = fd.getAll('%sKey');
            let vals = fd.getAll('%sValue');
			data = keys.map((k, i) => ({Key: k, Value: vals[i]}));
            $el.value = JSON.stringify(data);
            form.querySelectorAll('[name=%sKey], [name=%sValue]').forEach(el => el.disabled = true);
        });
	`, e.Name, e.Name, e.Name, e.Name)),
)
	return Div(Class(e.Classes), Group(nodes), finalInput)
}

func (e InputKeyValue) Parse(v any, ctx context.Context) (any, error) {
	vals, _ := v.([]string)
	if len(vals) == 0 || vals[0] == "" {
		return nil, fmt.Errorf("No value provided")
	}
	var dbJson datatypes.JSON
	err := dbJson.UnmarshalJSON([]byte(vals[0]))
	return dbJson, err
}

func (e InputKeyValue) GetName() string {
	return e.Name
}
