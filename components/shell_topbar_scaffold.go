package components

import (
	"context"

	. "maragu.dev/gomponents"
)

type ShellTopbarScaffold struct {
	Page
	Children  []PageInterface
	ExtraHead []PageInterface
}

func (e ShellTopbarScaffold) Body(ctx context.Context) Node {
	return ShellBase{
		ExtraHead: e.ExtraHead,
		Children: []PageInterface{
			LayoutTopbar{
				Children: e.Children,
			},
		},
	}.Body(ctx)
}

func (e ShellTopbarScaffold) Build(ctx context.Context) Node {
	return Render(ShellBase{
		ExtraHead: e.ExtraHead,
		Children: []PageInterface{
			LayoutTopbar{
				Children: e.Children,
			},
		},
	}, ctx)
}

func (e ShellTopbarScaffold) GetKey() string {
	return e.Key
}

func (e ShellTopbarScaffold) GetRoles() []string {
	return e.Roles
}

func (e ShellTopbarScaffold) GetChildren() []PageInterface {
	return e.Children
}

func (e *ShellTopbarScaffold) SetChildren(children []PageInterface) {
	e.Children = children
}
