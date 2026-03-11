package components

import (
	"context"
	"fmt"

	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

type FieldTitle struct {
	Page
	Getter  Getter
	Classes string
}

func (e FieldTitle) Build(ctx context.Context) Node {
	value := fmt.Sprintf("%s", IfOrGetter(e.Getter, ctx, ""))
	return Div(Class(fmt.Sprintf("text-xl font-semibold text-primary %s", e.Classes)), Text(value))
}
