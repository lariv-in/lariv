package components

import (
	"context"

	. "maragu.dev/gomponents"
)

type ShellScaffold struct {
	Sidebar  []PageInterface
	Children []PageInterface
}

func (e ShellScaffold) Body(ctx context.Context) Node {
	return ShellBase{
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
	}.Body(ctx)
}

func (e ShellScaffold) Build(ctx context.Context) Node {
	return ShellBase{
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

func (e ShellScaffold) GetChildren() []PageInterface {
	return append(e.Sidebar, e.Children...)
}
