package components

import (
	"context"

	"github.com/lariv-in/getters"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

type ButtonModal struct {
	Page
	Label       string
	Url         getters.Getter[string]
	Icon        string
	IconClasses string
	Classes     string
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
		if v, err := e.Url(ctx); err == nil {
			url = v
		}
	}

	content := Group{}
	if e.Icon != "" {
		content = append(content, Render(Icon{Name: e.Icon, Classes: e.IconClasses}, ctx))
	}
	if e.Label != "" {
		content = append(content, Text(e.Label))
	}

	buttonClasses := "btn w-full " + e.Classes
	if e.Icon != "" && e.Label != "" {
		buttonClasses += " inline-flex items-center gap-2"
	}

	return Div(Class("w-full"),
		Button(
			Type("button"),
			Class(buttonClasses),
			Attr("hx-get", url),
			Attr("hx-target", "next .modal-container"),
			Attr("hx-swap", "innerHTML"),
			Attr("hx-push-url", "false"),
			content,
		),
		Div(Class("modal-container")),
	)
}
