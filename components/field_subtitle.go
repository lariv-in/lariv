package components

import (
	"context"
	"log/slog"

	"github.com/lariv-in/lago/getters"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

type FieldSubtitle struct {
	Page
	Getter getters.Getter[string]
}

func (e FieldSubtitle) GetKey() string {
	return e.Key
}

func (e FieldSubtitle) GetRoles() []string {
	return e.Roles
}

func (e FieldSubtitle) Build(ctx context.Context) Node {
	value := ""
	if e.Getter != nil {
		v, err := e.Getter(ctx)
		if err != nil {
			slog.Error("FieldSubtitle getter failed", "error", err, "key", e.Key)
			return ContainerError{Error: getters.Static(err)}.Build(ctx)
		}
		value = v
	}
	return Div(Class("text-md text-gray-500"), Text(value))
}
