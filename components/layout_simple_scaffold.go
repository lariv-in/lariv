package components

import (
	"context"

	. "maragu.dev/gomponents"
)

type LayoutSimpleScaffold struct {
	Children []PageInterface
}

func (e LayoutSimpleScaffold) Build(ctx context.Context) Node {
	return LayoutBase{
		Children: []PageInterface{
			LayoutSimple{
				Children: e.Children,
			},
		},
	}.Build(ctx)
}

func (e LayoutSimpleScaffold) GetChildren() []PageInterface {
	return e.Children
}
