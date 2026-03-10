package components

import (
	"context"

	"maragu.dev/gomponents"
)

type LayoutScaffold struct {
	Sidebar  []PageInterface
	Children []PageInterface
}

func (e LayoutScaffold) Build(ctx context.Context) gomponents.Node {
	return LayoutBase{
		Children: []PageInterface{
			LayoutTopbar{
				Children: []PageInterface{
					LayoutSidebar{
						Sidebar:  e.Sidebar,
						Children: e.Children,
					},
				},
			},
		},
	}.Build(ctx)
}

func (e LayoutScaffold) GetChildren() []PageInterface {
	return append(e.Sidebar, e.Children...)
}
