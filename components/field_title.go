package components

import (
	"context"
	"fmt"

	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

type FieldTitle struct {
	Getter Getter
}

func (e FieldTitle) Build(ctx context.Context) Node {
	value := fmt.Sprintf("%s", IfOrGetter(e.Getter, ctx, ""))
	return Div(Text(value))
}
