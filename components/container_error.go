package components

import (
	"context"

	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

type ContainerError struct {
	Children []PageInterface
	Error    Getter
}

func (e ContainerError) Build(ctx context.Context) Node {
	group := Group{}
	for _, child := range e.Children {
		group = append(group, child.Build(ctx))
	}

	var errorNode Node
	if e.Error != nil {
		err := e.Error(ctx)
		if err != nil {
			errorNode = Span(Class("text-sm text-error"), Text(err.(error).Error()))
		}
	}

	return Div(Class("flex flex-col gap-1 w-full"), group, errorNode)
}

func (e ContainerError) GetChildren() []PageInterface {
	return e.Children
}
