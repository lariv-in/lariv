package components

import (
	"context"
	"log/slog"

	"github.com/lariv-in/lago/getters"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

type FieldCheckbox struct {
	Page
	Getter getters.Getter[bool]
}

func (e FieldCheckbox) GetKey() string {
	return e.Key
}

func (e FieldCheckbox) GetRoles() []string {
	return e.Roles
}

func (e FieldCheckbox) Build(ctx context.Context) Node {
	truthy := false
	if e.Getter != nil {
		value, err := e.Getter(ctx)
		if err != nil {
			slog.Error("FieldCheckbox getter failed", "error", err, "key", e.Key)
			return ContainerError{Error: getters.Static(err)}.Build(ctx)
		}
		truthy = value
	}

	if truthy {
		return Span(Render(Icon{Name: "check-circle", Classes: "text-success"}, ctx))
	}
	return Span(Render(Icon{Name: "x-circle", Classes: "text-error"}, ctx))
}
