package components

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/registry"
	"gorm.io/datatypes"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

type InputKeyValue struct {
	Page
	Getter  getters.Getter[datatypes.JSON]
	Keys    getters.Getter[[]string]
	Classes string
	Name    string
}

func (e InputKeyValue) GetKey() string {
	return e.Key
}

func (e InputKeyValue) GetRoles() []string {
	return e.Roles
}

func (e InputKeyValue) Build(ctx context.Context) Node {
	var val []registry.Pair[string, string]
	if e.Getter != nil {
		jsonData, err := e.Getter(ctx)
		if err != nil {
			slog.Error("InputKeyValue getter failed", "error", err, "key", e.Key)
		} else {
			if len(jsonData) > 0 {
				if err := json.Unmarshal(jsonData, &val); err != nil {
					slog.Error("InputKeyValue unmarshal failed", "error", err, "key", e.Key)
				}
			}
		}
	}

	if e.Keys == nil {
		slog.Error("InputKeyValue Keys is nil", "key", e.Key)
		return Div(Class(e.Classes))
	}
	keys, err := e.Keys(ctx)
	if err != nil {
		slog.Error("InputKeyValue Keys getter failed", "error", err, "key", e.Key)
		return Div(Class(e.Classes))
	}

	var nodes []Node
	for i, k := range keys {
		displayVal := ""
		if i < len(val) && val[i].Key == k {
			displayVal = val[i].Value
		}
		nodes = append(nodes,
			InputText{Hidden: true, Name: e.Name + "Key", Getter: getters.Static(k)}.Build(ctx),
			InputTextarea{Name: e.Name + "Value", Label: k, Getter: getters.Static(displayVal)}.Build(ctx),
		)
	}

	finalInput := Input(
		Type("hidden"),
		Name(e.Name),
		Attr("x-data"),
		Attr("x-init", fmt.Sprintf(`
	$el.closest('form').addEventListener('submit', (e) => {
		let form = e.currentTarget;
		let data = [];
		let fd = new FormData(form);
            let keys = fd.getAll('%sKey');
            let vals = fd.getAll('%sValue');
			data = keys.map((k, i) => ({Key: k, Value: vals[i]}));
            $el.value = JSON.stringify(data);
            form.querySelectorAll('[name=%sKey], [name=%sValue]').forEach(el => el.disabled = true);
        }, true);
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
