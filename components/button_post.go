package components

import (
	"context"
	"fmt"
	"net/http"

	"github.com/lariv-in/getters"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

type ButtonPost struct {
	Page
	Label   string
	URL     getters.Getter
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
		if val := e.URL(ctx); val != nil {
			url = fmt.Sprintf("%s", val)
		}
	}
	return Form(
		Action(url), Method(http.MethodPost),
		Button(Type("submit"), Class(fmt.Sprintf("btn w-full %s", e.Classes)), Text(e.Label)),
	)
}
