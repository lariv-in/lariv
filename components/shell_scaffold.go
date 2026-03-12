package components

import (
	"context"

	. "maragu.dev/gomponents"
)

type ShellScaffold struct {
	Page
	Sidebar  []PageInterface
	Children []PageInterface
	ExtraHead []PageInterface
}

func (e ShellScaffold) Body(ctx context.Context) Node {
	return ShellBase{
		ExtraHead: e.ExtraHead,
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
	return Render(ShellBase{
		ExtraHead: e.ExtraHead,
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
	}, ctx)
}

func (e ShellScaffold) GetChildren() []PageInterface {
	return append(e.Sidebar, e.Children...)
}
