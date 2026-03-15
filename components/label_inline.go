package components

import (
	"context"
	"fmt"

	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

type LabelInline struct {
	Page
	Title    string
	Children []PageInterface
	Classes  string
}

func (e LabelInline) Build(ctx context.Context) Node {
	var childNodes []Node
	for _, child := range e.Children {
		childNodes = append(childNodes, Render(child, ctx))
	}
	return Div(Class(fmt.Sprintf("flex gap-2 %s", e.Classes)),
		Span(Class("text-primary font-bold"), Text(e.Title+":")),
		Group(childNodes),
	)
}

func (e LabelInline) GetKey() string {
	return e.Key
}

func (e LabelInline) GetRoles() []string {
	return e.Roles
}

func (e LabelInline) GetChildren() []PageInterface {
	return e.Children
}

func (e *LabelInline) SetChildren(children []PageInterface) {
	e.Children = children
}
