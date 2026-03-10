package components

import (
	"context"

	. "maragu.dev/gomponents"
)

type ShellAuthScaffold struct {
	Children []PageInterface
}

func (e ShellAuthScaffold) Body(ctx context.Context) Node {
	return ShellBase{
		Children: []PageInterface{LayoutCard{e.Children}},
	}.Body(ctx)
}

func (e ShellAuthScaffold) Build(ctx context.Context) Node {
	return ShellBase{
		Children: []PageInterface{LayoutCard{e.Children}},
	}.Build(ctx)
}

func (e ShellAuthScaffold) GetChildren() []PageInterface {
	return e.Children
}
