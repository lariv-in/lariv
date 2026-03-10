package components

import (
	"context"

	. "maragu.dev/gomponents"
)

type ShellSimpleScaffold struct {
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
	return ShellBase{
		Children: []PageInterface{
			LayoutSimple{
				Children: e.Children,
			},
		},
	}.Build(ctx)
}

func (e ShellSimpleScaffold) GetChildren() []PageInterface {
	return e.Children
}
