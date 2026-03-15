package components

import (
	"context"

	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

type LayoutCard struct {
	Page
	Children []PageInterface
}

func (e LayoutCard) Build(ctx context.Context) Node {
	group := Group{}
	for _, child := range e.Children {
		group = append(group, Render(child, ctx))
	}
	return Div(Class("min-h-screen flex items-center justify-center bg-base-200"),
		Progress(Class("progress w-full fixed top-0 left-0 h-1 z-50"), ID("global-loading-indicator")),
		Div(Class("card shadow-xl"), Div(Class("card-body"),
			group,
		)),
	)
}

func (e LayoutCard) GetKey() string {
	return e.Key
}

func (e LayoutCard) GetRoles() []string {
	return e.Roles
}

func (e LayoutCard) GetChildren() []PageInterface {
	return e.Children
}

func (e *LayoutCard) SetChildren(children []PageInterface) {
	e.Children = children
}
