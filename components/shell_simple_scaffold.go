package components

import (
	"context"

	. "maragu.dev/gomponents"
)

type ShellSimpleScaffold struct {
	Page
	Children  []PageInterface
	ExtraHead []PageInterface
}

func (e ShellSimpleScaffold) Body(ctx context.Context) Node {
	return ShellBase{
		ExtraHead: e.ExtraHead,
		Children: []PageInterface{
			LayoutSimple{
				Children: e.Children,
			},
		},
	}.Body(ctx)
}

func (e ShellSimpleScaffold) Build(ctx context.Context) Node {
	return Render(ShellBase{
		ExtraHead: e.ExtraHead,
		Children: []PageInterface{
			LayoutSimple{
				Children: e.Children,
			},
		},
	}, ctx)
}

func (e ShellSimpleScaffold) GetKey() string {
	return e.Key
}

func (e ShellSimpleScaffold) GetRoles() []string {
	return e.Roles
}

func (e ShellSimpleScaffold) GetChildren() []PageInterface {
	return e.Children
}

func (e *ShellSimpleScaffold) SetChildren(children []PageInterface) {
	e.Children = children
}
