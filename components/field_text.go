package components

import (
	"context"
	"fmt"

	"github.com/lariv-in/getters"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

type FieldText struct {
	Page
	Getter  getters.Getter
	Classes string
}

func (e FieldText) Build(ctx context.Context) Node {
	value := fmt.Sprintf("%s", getters.IfOrGetter(e.Getter, ctx, ""))
	return Div(Class(e.Classes), Text(value))
}
