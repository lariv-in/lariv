package components

import (
	"context"
	"fmt"

	"maragu.dev/gomponents"
	"maragu.dev/gomponents/html"
)

type ContainerColumn struct {
	Page
	Children []PageInterface
	Classes  string
}

func (e ContainerColumn) Build(ctx context.Context) gomponents.Node {
	group := gomponents.Group{}
	for _, child := range e.Children {
		group = append(group, Render(child, ctx))
	}
	return html.Div(html.Class(fmt.Sprintf("flex flex-col gap-1 %s", e.Classes)), group)
}

func (e ContainerColumn) GetKey() string {
	return e.Key
}

func (e ContainerColumn) GetRoles() []string {
	return e.Roles
}

func (e ContainerColumn) GetChildren() []PageInterface {
	return e.Children
}

func (e *ContainerColumn) SetChildren(children []PageInterface) {
	e.Children = children
}
