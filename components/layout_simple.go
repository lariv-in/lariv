package components

import (
	"context"

	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

type LayoutSimple struct {
	Children []PageInterface
}

func (e LayoutSimple) Build(ctx context.Context) Node {
	return ContainerHtml{
		Children: e.Children,
		Html: func(children Node) Node {
			return Div(Class("size-full overflow-y-auto p-4"),
				children,
			)
		},
	}.Build(ctx)
}

func (e LayoutSimple) GetChildren() []PageInterface {
	return e.Children
}
