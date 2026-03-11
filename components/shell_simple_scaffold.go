package components

import (
	"context"

	. "maragu.dev/gomponents"
)

type ShellSimpleScaffold struct {
	Page
	Children []PageInterface
}

func (e ShellSimpleScaffold) Body(ctx context.Context) Node {
	return ShellBase{
		Children: []PageInterface{
			LayoutSimple{
				Children: e.Children,
			},
		},
	}.Body(ctx)
}

func (e ShellSimpleScaffold) Build(ctx context.Context) Node {
	return Render(ShellBase{
		Children: []PageInterface{
			LayoutSimple{
				Children: e.Children,
			},
		},
	}, ctx)
}

func (e ShellSimpleScaffold) GetChildren() []PageInterface {
	return e.Children
}
