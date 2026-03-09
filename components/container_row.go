package components

import (
	"context"
	"fmt"

	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

type ContainerRow struct {
	Children []PageInterface
	Classes  string
}

func (e ContainerRow) Build(ctx context.Context) Node {
	group := Group{}
	for _, child := range e.Children {
		group = append(group, child.Build(ctx))
	}
	return Div(Class(fmt.Sprintf("flex flex-row gap-1 %s", e.Classes)), group)
}


func (e ContainerRow) GetChildren() []PageInterface {
	return e.Children
}
