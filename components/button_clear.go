package components

import (
	"context"

	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

type ButtonClear struct {
	Page
	Label   string
	Classes string
}

func (e ButtonClear) GetKey() string {
	return e.Key
}

func (e ButtonClear) GetRoles() []string {
	return e.Roles
}

func (e ButtonClear) Build(ctx context.Context) Node {
	label := e.Label
	if label == "" {
		label = "Clear"
	}
	return Button(Type("button"), Class("btn btn-ghost my-2 "+e.Classes), Text(label),
		Attr("onclick", "this.closest('form').querySelectorAll('input,select,textarea').forEach(el => { el.value = ''; });"),
	)
}
