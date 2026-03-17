package components

import (
	"context"

	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

type LayoutSimple struct {
	Page
	Children []PageInterface
}

func (e LayoutSimple) Build(ctx context.Context) Node {
	return Render(ContainerHTML{
		Children: e.Children,
		HTML: func(ctx context.Context, children Node) Node {
			return Div(Class("size-full overflow-y-auto p-4"),
				children,
			)
		},
	}, ctx)
}

func (e LayoutSimple) GetKey() string {
	return e.Key
}

func (e LayoutSimple) GetRoles() []string {
	return e.Roles
}

func (e LayoutSimple) GetChildren() []PageInterface {
	return e.Children
}

func (e *LayoutSimple) SetChildren(children []PageInterface) {
	e.Children = children
}
