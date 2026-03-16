package components

import (
	"context"
	"net/http"

	"github.com/lariv-in/getters"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

type ButtonPost struct {
	Page
	Label   string
	URL     getters.Getter[string]
	Classes string
}

func (e ButtonPost) GetKey() string {
	return e.Key
}

func (e ButtonPost) GetRoles() []string {
	return e.Roles
}

func (e ButtonPost) Build(ctx context.Context) Node {
	url := ""
	if e.URL != nil {
		if v, err := e.URL(ctx); err == nil {
			url = v
		}
	}
	return Form(
		Action(url), Method(http.MethodPost),
		Button(Type("submit"), Class("btn w-full "+e.Classes), Text(e.Label)),
	)
}
