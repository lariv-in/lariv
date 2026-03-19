package components

import (
	"context"

	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

// ClientData wraps children with Alpine x-data/x-init state.
type ClientData struct {
	Page
	Data     string
	Init     string
	Children []PageInterface
}

func (e ClientData) Build(ctx context.Context) Node {
	data := e.Data
	if data == "" {
		data = "{}"
	}

	group := Group{}
	for _, child := range e.Children {
		group = append(group, Render(child, ctx))
	}

	return Div(
		Attr("x-data", data),
		If(e.Init != "", Attr("x-init", e.Init)),
		group,
	)
}

func (e ClientData) GetKey() string {
	return e.Key
}

func (e ClientData) GetRoles() []string {
	return e.Roles
}

func (e ClientData) GetChildren() []PageInterface {
	return e.Children
}

func (e *ClientData) SetChildren(children []PageInterface) {
	e.Children = children
}
