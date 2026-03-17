package components

import (
	"context"

	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

type ButtonClear struct {
	Page
	Label       string
	Icon        string
	IconClasses string
	Classes     string
}

func (e ButtonClear) GetKey() string {
	return e.Key
}

func (e ButtonClear) GetRoles() []string {
	return e.Roles
}

func (e ButtonClear) Build(ctx context.Context) Node {
	label := e.Label
	if label == "" {
		label = "Clear"
	}
	content := Group{}
	if e.Icon != "" {
		content = append(content, Render(Icon{Name: e.Icon, Classes: e.IconClasses}, ctx))
	}
	if label != "" {
		content = append(content, Text(label))
	}

	classes := "btn btn-ghost my-2 " + e.Classes
	if e.Icon != "" && label != "" {
		classes += " inline-flex items-center gap-2"
	}

	return Button(Type("button"), Class(classes), content,
		Attr("onclick", "this.closest('form').querySelectorAll('input,select,textarea').forEach(el => { el.value = ''; });"),
	)
}
