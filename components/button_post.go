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
	Url     getters.Getter
	Classes string
}

func (e ButtonPost) Build(ctx context.Context) Node {
	url := ""
	if e.Url != nil {
		if val := e.Url(ctx); val != nil {
			url = fmt.Sprintf("%s", val)
		}
	}
	return Form(
		Action(url), Method(http.MethodPost),
		Button(Type("submit"), Class(fmt.Sprintf("btn w-full %s", e.Classes)), Text(e.Label)),
	)
}
