package components

import (
	"context"
	"html/template"
	"io"

	"github.com/lariv-in/lariv/components"
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

func (e ThemeButton) Build(cat components.Catalog, ctx context.Context, w io.Writer) error {
	sunIcon, err := components.RenderHTML(components.Icon{Name: "sun"}, cat, ctx)
	if err != nil {
		return err
	}
	moonIcon, err := components.RenderHTML(components.Icon{Name: "moon"}, cat, ctx)
	if err != nil {
		return err
	}
	return execute(w, "theme_button", struct {
		Classes  string
		SunIcon  template.HTML
		MoonIcon template.HTML
	}{Classes: e.Classes, SunIcon: sunIcon, MoonIcon: moonIcon})
}
