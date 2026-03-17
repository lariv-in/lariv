package components

import (
	"context"

	"github.com/lariv-in/getters"
	"maragu.dev/gomponents"
	"maragu.dev/gomponents/html"
)

type ButtonLink struct {
	Page
	Label       string
	GetterLabel getters.Getter[string]
	Link        getters.Getter[string]
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
	return html.A(html.Href(link), html.Class("btn "+e.Classes), gomponents.Text(label))
}
