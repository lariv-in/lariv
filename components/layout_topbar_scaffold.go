package components

import (
	"context"

	. "maragu.dev/gomponents"
)

type LayoutTopbarScaffold struct {
	Children []PageInterface
}

func (e LayoutTopbarScaffold) Build(ctx context.Context) Node {
	return LayoutBase{
		Children: []PageInterface{
			LayoutTopbar{
				Children: e.Children,
			},
		},
	}.Build(ctx)
}

func (e LayoutTopbarScaffold) GetChildren() []PageInterface {
	return e.Children
}
