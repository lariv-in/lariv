package components

import (
	"context"

	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

type ButtonSubmit struct {
	Page
	Label       string
	Icon        string
	IconClasses string
	Classes     string
}

func (e ButtonSubmit) GetKey() string {
	return e.Key
}

func (e ButtonSubmit) GetRoles() []string {
	return e.Roles
}

func (e ButtonSubmit) Build(ctx context.Context) Node {
	content := Group{}
	if e.Icon != "" {
		content = append(content, Render(Icon{Name: e.Icon, Classes: e.IconClasses}, ctx))
	}
	if e.Label != "" {
		content = append(content, Text(e.Label))
	}

	classes := "btn btn-primary " + e.Classes
	if e.Icon != "" && e.Label != "" {
		classes += " inline-flex items-center gap-2"
	}

	return Button(Type("submit"), Class(classes), content)
}
