package components

import (
	"context"
	"fmt"

	"github.com/lariv-in/getters"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

type FieldTextArea struct {
	Page
	Getter  getters.Getter
	Classes string
}

func (e FieldTextArea) Build(ctx context.Context) Node {
	value := fmt.Sprintf("%s", getters.IfOrGetter(e.Getter, ctx, ""))
	return Div(Class(fmt.Sprintf("%s whitespace-pre-wrap", e.Classes)), Text(value))
}
