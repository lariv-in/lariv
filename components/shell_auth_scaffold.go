package components

import (
	"context"

	. "maragu.dev/gomponents"
)

type ShellAuthScaffold struct {
	Page
	Children  []PageInterface
	ExtraHead []PageInterface
}

func (e ShellAuthScaffold) Body(ctx context.Context) Node {
	return ShellBase{
		ExtraHead: e.ExtraHead,
		Children:  []PageInterface{LayoutCard{Page{}, e.Children}},
	}.Body(ctx)
}

func (e ShellAuthScaffold) Build(ctx context.Context) Node {
	return Render(ShellBase{
		ExtraHead: e.ExtraHead,
		Children:  []PageInterface{LayoutCard{Page{}, e.Children}},
	}, ctx)
}

func (e ShellAuthScaffold) GetKey() string {
	return e.Key
}

func (e ShellAuthScaffold) GetRoles() []string {
	return e.Roles
}

func (e ShellAuthScaffold) GetChildren() []PageInterface {
	return e.Children
}

func (e *ShellAuthScaffold) SetChildren(children []PageInterface) {
	e.Children = children
}
