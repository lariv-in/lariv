package components

import (
	"context"
	"fmt"

	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

type InputClear struct {
	Label   string
	Classes string
}

func (e InputClear) Build(ctx context.Context) Node {
	label := e.Label
	if label == "" {
		label = "Clear"
	}
	return Button(Type("button"), Class(fmt.Sprintf("btn btn-ghost my-2 %s", e.Classes)), Text(label),
		Attr("onclick", "this.closest('form').querySelectorAll('input,select,textarea').forEach(el => { el.value = ''; });"),
	)
}
