package components

import (
	"context"

	"github.com/lariv-in/lago/getters"
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
	Attr        getters.Getter[Node]
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

	buttonClasses := "btn " + e.Classes
	if e.Icon != "" && e.Label != "" {
		buttonClasses += " inline-flex items-center gap-2"
	}

	buttonAttrs := []Node{
		Type("button"),
		Class(buttonClasses),
		Attr("hx-get", url),
		Attr("hx-target", HTMXTargetBodyModal),
		Attr("hx-swap", HTMXSwapBodyModal),
		Attr("hx-push-url", "false"),
	}
	if e.Attr != nil {
		extra, err := e.Attr(ctx)
		if err != nil {
			return ContainerError{Error: getters.Static(err)}.Build(ctx)
		}
		if extra != nil {
			buttonAttrs = append(buttonAttrs, extra)
		}
	}
	buttonAttrs = append(buttonAttrs, content)

	return Div(Class("w-full fk-modal-host"),
		Button(Group(buttonAttrs)),
	)
}
