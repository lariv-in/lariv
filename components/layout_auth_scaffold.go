package components

import (
	"context"
	. "maragu.dev/gomponents"
)

type LayoutAuthScaffold struct {
	Children []PageInterface
}

func (e LayoutAuthScaffold) Build(ctx context.Context) Node {
	return Node(LayoutBase{
		Children: []PageInterface{LayoutCard{e.Children}},
	}.Build(ctx))
}


func (e LayoutAuthScaffold) GetChildren() []PageInterface {
	return e.Children
}
