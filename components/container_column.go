package components

import (
	"context"
	"fmt"

	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

type ContainerColumn struct {
	Children []PageInterface
	Classes  string
}

func (e ContainerColumn) Build(ctx context.Context) Node {
	group := Group{}
	for _, child := range e.Children {
		group = append(group, child.Build(ctx))
	}
	return Div(Class(fmt.Sprintf("flex flex-col gap-1 %s", e.Classes)), group)
}


func (e ContainerColumn) GetChildren() []PageInterface {
	return e.Children
}
