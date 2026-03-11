package components

import (
	"context"
	"fmt"

	"github.com/lariv-in/getters"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

type FieldSubtitle struct {
	Page
	Getter getters.Getter
}

func (e FieldSubtitle) Build(ctx context.Context) Node {
	value := fmt.Sprintf("%s", getters.IfOrGetter(e.Getter, ctx, ""))
	return Div(Class("text-md text-gray-500"), Text(value))
}
