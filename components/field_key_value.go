package components

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/registry"
	"gorm.io/datatypes"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

type FieldKeyValue struct {
	Page
	Getter  getters.Getter[datatypes.JSON]
	Classes string
}

func (e FieldKeyValue) GetKey() string {
	return e.Key
}

func (e FieldKeyValue) GetRoles() []string {
	return e.Roles
}

func (e FieldKeyValue) Build(ctx context.Context) Node {
	if e.Getter == nil {
		return Div()
	}

	jsonData, err := e.Getter(ctx)
	if err != nil {
		slog.Error("FieldKeyValue getter failed", "error", err, "key", e.Key)
		return ContainerError{Error: getters.Static(err)}.Build(ctx)
	}
	if len(jsonData) == 0 {
		return Div()
	}

	var val []registry.Pair[string, string]
	err = json.Unmarshal(jsonData, &val)
	if err != nil {
		slog.Error("FieldKeyValue unmarshal failed", "error", err, "key", e.Key)
		return ContainerError{Error: getters.Static(err)}.Build(ctx)
	}

	var nodes []Node
	for _, r := range val {
		nodes = append(nodes,
			Div(Class("mb-4 pb-4 border-b border-base-300 last:border-b-0"),
				Div(Class("font-medium text-sm text-base-content/70 mb-1"), Text(r.Key)),
				Div(Class("whitespace-pre-wrap"), Text(r.Value)),
			),
		)
	}
	return Div(Class(e.Classes), Group(nodes))
}
