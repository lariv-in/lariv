package components

import (
	"context"

	"github.com/lariv-in/lago/components"
	"maragu.dev/gomponents"
	"maragu.dev/gomponents/html"
)

type ThemeButton struct {
	components.Page
	Classes string
}

func (e ThemeButton) GetKey() string {
	return e.Key
}

func (e ThemeButton) GetRoles() []string {
	return e.Roles
}

func (e ThemeButton) Build(ctx context.Context) gomponents.Node {
	return html.Button(
		html.Type("button"),
		html.Class("btn items-center "+e.Classes),
		gomponents.Attr("@click", "toggleTheme()"),
		html.Span(
			html.Class("inline-flex items-center justify-center"),
			gomponents.Attr("x-show", "theme === 'light'"),
			components.Render(components.Icon{Name: "sun"}, ctx),
		),
		html.Span(
			html.Class("inline-flex items-center justify-center"),
			gomponents.Attr("x-show", "theme !== 'light'"),
			components.Render(components.Icon{Name: "moon"}, ctx),
		),
	)
}
