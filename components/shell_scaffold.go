package components

import (
	"context"

	. "maragu.dev/gomponents"
)

type ShellScaffold struct {
	Page
	Sidebar   []PageInterface
	Children  []PageInterface
	ExtraHead []PageInterface
}

func (e ShellScaffold) Body(ctx context.Context) Node {
	return ShellBase{
		ExtraHead: e.ExtraHead,
		Children: []PageInterface{
			LayoutTopbar{
				Children: []PageInterface{
					LayoutSidebar{
						Sidebar:  e.Sidebar,
						Children: e.Children,
					},
				},
			},
		},
	}.Body(ctx)
}

func (e ShellScaffold) Build(ctx context.Context) Node {
	return Render(ShellBase{
		ExtraHead: e.ExtraHead,
		Children: []PageInterface{
			LayoutTopbar{
				Children: []PageInterface{
					LayoutSidebar{
						Sidebar:  e.Sidebar,
						Children: e.Children,
					},
				},
			},
		},
	}, ctx)
}

func (e ShellScaffold) GetKey() string {
	return e.Key
}

func (e ShellScaffold) GetRoles() []string {
	return e.Roles
}

func (e ShellScaffold) GetChildren() []PageInterface {
	return append(e.Sidebar, e.Children...)
}

func (e *ShellScaffold) SetChildren(children []PageInterface) {
	offset := 0
	nSidebar := len(e.Sidebar)
	end := min(offset+nSidebar, len(children))
	e.Sidebar = children[offset:end]
	offset = end
	if offset >= len(children) {
		return
	}
	nContent := len(e.Children)
	end = min(offset+nContent, len(children))
	e.Children = children[offset:end]
	offset = end
	if offset < len(children) {
		e.Children = append(e.Children, children[offset:]...)
	}
}
