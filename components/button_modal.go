package components

import (
	"context"
	"fmt"

	"github.com/lariv-in/getters"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

type ButtonModal struct {
	Page
	Label   string
	Url     getters.Getter
	Classes string
}

func (e ButtonModal) GetKey() string {
	return e.Key
}

func (e ButtonModal) GetRoles() []string {
	return e.Roles
}

func (e ButtonModal) Build(ctx context.Context) Node {
	url := ""
	if e.Url != nil {
		if val := e.Url(ctx); val != nil {
			url = fmt.Sprintf("%s", val)
		}
	}
	return Div(Class("w-full"),
		Button(
			Type("button"),
			Class(fmt.Sprintf("btn w-full %s", e.Classes)),
			Attr("hx-get", url),
			Attr("hx-target", "next .modal-container"),
			Attr("hx-swap", "innerHTML"),
			Attr("hx-push-url", "false"),
			Text(e.Label),
		),
		Div(Class("modal-container")),
	)
}
