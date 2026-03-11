package components

import (
	"context"

	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

type FieldCheckbox struct {
	Page
	Getter Getter
}

func (e FieldCheckbox) Build(ctx context.Context) Node {
	value := IfOrGetter(e.Getter, ctx, false)
	truthy := false
	if b, ok := value.(bool); ok {
		truthy = b
	}

	if truthy {
		return Span(Render(Icon{Name: "check-circle", Classes: "text-success"}, ctx))
	}
	return Span(Render(Icon{Name: "x-circle", Classes: "text-error"}, ctx))
}
