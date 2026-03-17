package components

import (
	"context"

	"maragu.dev/gomponents"
)

type ContainerHTML struct {
	Page
	Children []PageInterface
	HTML     func(context.Context, gomponents.Node) gomponents.Node
}

func (e ContainerHTML) Build(ctx context.Context) gomponents.Node {
	group := gomponents.Group{}
	for _, child := range e.Children {
		group = append(group, Render(child, ctx))
	}
	if e.HTML != nil {
		return e.HTML(ctx, group)
	}
	return group
}

func (e ContainerHTML) GetKey() string {
	return e.Key
}

func (e ContainerHTML) GetRoles() []string {
	return e.Roles
}

func (e ContainerHTML) GetChildren() []PageInterface {
	return e.Children
}

func (e *ContainerHTML) SetChildren(children []PageInterface) {
	e.Children = children
}
