package components

import (
	"context"

	. "maragu.dev/gomponents"
)

type ContainerHtml struct {
	Children []PageInterface
	Html     func(Node) Node
}

func (e ContainerHtml) Build(ctx context.Context) Node {
	group := Group{}
	for _, child := range e.Children {
		group = append(group, child.Build(ctx))
	}
	if e.Html != nil {
		return e.Html(group)
	}
	return group
}

func (e ContainerHtml) GetChildren() []PageInterface {
	return e.Children
}
