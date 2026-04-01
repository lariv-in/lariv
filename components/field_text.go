package components

import (
	"context"
	"log/slog"

	"github.com/lariv-in/lago/getters"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

type FieldText struct {
	Page
	Getter  getters.Getter[string]
	Classes string
}

func (e FieldText) GetKey() string {
	return e.Key
}

func (e FieldText) GetRoles() []string {
	return e.Roles
}

func (e FieldText) Build(ctx context.Context) Node {
	value := ""
	if e.Getter != nil {
		v, err := e.Getter(ctx)
		if err != nil {
			slog.Error("FieldText getter failed", "error", err, "key", e.Key)
			return ContainerError{Error: getters.Static(err)}.Build(ctx)
		}
		value = v
	}
	return Div(Class(e.Classes), Text(value))
}
