package components

import (
	"context"

	. "maragu.dev/gomponents"
)

type ShellTopbarScaffold struct {
	Children []PageInterface
}

func (e ShellTopbarScaffold) Body(ctx context.Context) Node {
	return ShellBase{
		Children: []PageInterface{
			LayoutTopbar{
				Children: e.Children,
			},
		},
	}.Body(ctx)
}

func (e ShellTopbarScaffold) Build(ctx context.Context) Node {
	return ShellBase{
		Children: []PageInterface{
			LayoutTopbar{
				Children: e.Children,
			},
		},
	}.Build(ctx)
}

func (e ShellTopbarScaffold) GetChildren() []PageInterface {
	return e.Children
}
