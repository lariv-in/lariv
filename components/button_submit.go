package components

import (
	"context"
	"fmt"

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
	return Button(Type("submit"), Class(fmt.Sprintf("btn btn-primary my-2 %s", e.Classes)), Text(e.Label))
}
