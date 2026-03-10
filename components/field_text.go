package components

import (
	"context"
	"fmt"

	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

type FieldText struct {
	Getter  Getter
	Classes string
}

func (e FieldText) Build(ctx context.Context) Node {
	value := fmt.Sprintf("%s", IfOrGetter(e.Getter, ctx, ""))
	return Div(Class(e.Classes), Text(value))
}
