package components

import (
	"context"

	"github.com/lariv-in/lago/getters"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

type ButtonDownload struct {
	Page
	Label       string
	Link        getters.Getter[string]
	Icon        string
	IconClasses string
	Classes     string
}

func (e ButtonDownload) GetKey() string {
	return e.Key
}

func (e ButtonDownload) GetRoles() []string {
	return e.Roles
}

func (e ButtonDownload) Build(ctx context.Context) Node {
	link := ""
	if e.Link != nil {
		if v, err := e.Link(ctx); err == nil {
			link = v
		}
	}
	content := Group{}
	if e.Icon != "" {
		content = append(content, Render(Icon{Name: e.Icon, Classes: e.IconClasses}, ctx))
	}
	if e.Label != "" {
		content = append(content, Text(e.Label))
	}

	classes := "btn " + e.Classes
	if e.Icon != "" && e.Label != "" {
		classes += " inline-flex gap-2"
	}

	return A(
		Href(link),
		Class(classes),
		Attr("data-hx-boost", "false"),
		Attr("download"),
		content,
	)
}
