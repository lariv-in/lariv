package components

import (
	"context"
	"fmt"

	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

type FieldSubtitle struct {
	Page
	Getter Getter
}

func (e FieldSubtitle) Build(ctx context.Context) Node {
	value := fmt.Sprintf("%s", IfOrGetter(e.Getter, ctx, ""))
	return Div(Class("text-md text-gray-500"), Text(value))
}
