package components

import (
	"context"

	"github.com/lariv-in/lago/getters"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

type HTMXPolling struct {
	Page
	URL      getters.Getter[string]
	Children []PageInterface
}

func (e HTMXPolling) Build(ctx context.Context) Node {
	var children Group
	for _, child := range e.Children {
		children = append(children, Render(child, ctx))
	}
	url, err := e.URL(ctx)
	if err != nil {
		return ContainerError{Error: getters.Static(err)}.Build(ctx)
	}
	return Div(
		Attr("hx-get", url),
		Attr("hx-target", "body"),
		Attr("hx-swap", "outerHTML"),
		Attr("hx-trigger", "every 2s"),
		children,
	)
}

func (e HTMXPolling) GetKey() string {
	return e.Key
}

func (e HTMXPolling) GetRoles() []string {
	return e.Roles
}

func (e HTMXPolling) GetChildren() []PageInterface {
	return e.Children
}

func (e *HTMXPolling) SetChildren(children []PageInterface) {
	e.Children = children
}
