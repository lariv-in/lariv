package components

import (
	"context"

	"github.com/lariv-in/lago/getters"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

type ContainerError struct {
	Page
	Children []PageInterface
	Error    getters.Getter[error]
}

func (e ContainerError) Build(ctx context.Context) Node {
	group := Group{}
	for _, child := range e.Children {
		group = append(group, Render(child, ctx))
	}

	var errorNode Node
	if e.Error != nil {
		err, _ := e.Error(ctx)
		if err != nil {
			errorNode = Span(Class("text-sm text-error"), Text(err.Error()))
		}
	}

	return Div(Class("flex flex-col gap-1 w-full"), group, errorNode)
}

func (e ContainerError) GetKey() string {
	return e.Key
}

func (e ContainerError) GetRoles() []string {
	return e.Roles
}

func (e ContainerError) GetChildren() []PageInterface {
	return e.Children
}

func (e *ContainerError) SetChildren(children []PageInterface) {
	e.Children = children
}
