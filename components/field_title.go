package components

import (
	"context"
	"log/slog"

	"github.com/lariv-in/lago/getters"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

type FieldTitle struct {
	Page
	Getter  getters.Getter[string]
	Classes string
}

func (e FieldTitle) GetKey() string {
	return e.Key
}

func (e FieldTitle) GetRoles() []string {
	return e.Roles
}

func (e FieldTitle) Build(ctx context.Context) Node {
	value := ""
	if e.Getter != nil {
		v, err := e.Getter(ctx)
		if err != nil {
			slog.Error("FieldTitle getter failed", "error", err, "key", e.Key)
			return ContainerError{Error: getters.Static(err)}.Build(ctx)
		}
		value = v
	}
	return Div(Class("text-xl font-semibold text-primary "+e.Classes), Text(value))
}
