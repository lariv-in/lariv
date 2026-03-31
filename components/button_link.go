package components

import (
	"context"

	"github.com/lariv-in/lago/getters"
	"maragu.dev/gomponents"
	"maragu.dev/gomponents/html"
)

type ButtonLink struct {
	Page
	Label       string
	GetterLabel getters.Getter[string]
	Link        getters.Getter[string]
	Icon        string
	IconClasses string
	Classes     string
}

func (e ButtonLink) GetKey() string {
	return e.Key
}

func (e ButtonLink) GetRoles() []string {
	return e.Roles
}

func (e ButtonLink) Build(ctx context.Context) gomponents.Node {
	link := ""
	if e.Link != nil {
		if v, err := e.Link(ctx); err == nil {
			link = v
		}
	}
	label := e.Label
	if e.GetterLabel != nil {
		if v, err := e.GetterLabel(ctx); err == nil {
			label = v
		}
	}

	content := gomponents.Group{}
	if e.Icon != "" {
		content = append(content, Render(Icon{Name: e.Icon, Classes: e.IconClasses}, ctx))
	}
	if label != "" {
		content = append(content, gomponents.Text(label))
	}

	classes := "btn " + e.Classes
	if e.Icon != "" && label != "" {
		classes += " flex items-center gap-2"
	}
	return html.A(html.Href(link), html.Class(classes), content)
}
