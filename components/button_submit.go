package components

import (
	"context"

	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

type ButtonSubmit struct {
	Page
	Label   string
	Classes string
}

func (e ButtonSubmit) GetKey() string {
	return e.Key
}

func (e ButtonSubmit) GetRoles() []string {
	return e.Roles
}

func (e ButtonSubmit) Build(ctx context.Context) Node {
	return Button(Type("submit"), Class("btn btn-primary my-2 "+e.Classes), Text(e.Label))
}
