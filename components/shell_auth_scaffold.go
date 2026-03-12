package components

import (
	"context"

	. "maragu.dev/gomponents"
)

type ShellAuthScaffold struct {
	Page
	Children []PageInterface
	ExtraHead []PageInterface
}

func (e ShellAuthScaffold) Body(ctx context.Context) Node {
	return ShellBase{
		ExtraHead: e.ExtraHead,
		Children: []PageInterface{LayoutCard{Page{}, e.Children}},
	}.Body(ctx)
}

func (e ShellAuthScaffold) Build(ctx context.Context) Node {
	return Render(ShellBase{
		ExtraHead: e.ExtraHead,
		Children: []PageInterface{LayoutCard{Page{}, e.Children}},
	}, ctx)
}

func (e ShellAuthScaffold) GetChildren() []PageInterface {
	return e.Children
}
