package components

import (
	"context"
	"log/slog"

	"github.com/lariv-in/lago/getters"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

type FieldTextArea struct {
	Page
	Getter  getters.Getter[string]
	Classes string
}

func (e FieldTextArea) GetKey() string {
	return e.Key
}

func (e FieldTextArea) GetRoles() []string {
	return e.Roles
}

func (e FieldTextArea) Build(ctx context.Context) Node {
	value := ""
	if e.Getter != nil {
		v, err := e.Getter(ctx)
		if err != nil {
			slog.Error("FieldTextArea getter failed", "error", err, "key", e.Key)
			return ContainerError{Error: getters.Static(err)}.Build(ctx)
		}
		value = v
	}
	return Div(Class(e.Classes+" whitespace-pre-wrap"), Text(value))
}
