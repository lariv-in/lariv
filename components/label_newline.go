package components

import (
	"context"
	"fmt"

	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

// LabelNewline is like LabelInline but renders Children on a new line below the title.
type LabelNewline struct {
	Page
	Title    string
	Children []PageInterface
	Classes  string
}

func (e LabelNewline) Build(ctx context.Context) Node {
	var childNodes []Node
	for _, child := range e.Children {
		childNodes = append(childNodes, Render(child, ctx))
	}
	return Div(Class(fmt.Sprintf("flex flex-col %s", e.Classes)),
		Span(Class("text-primary font-bold"), Text(e.Title+":")),
		Group(childNodes),
	)
}

func (e LabelNewline) GetKey() string {
	return e.Key
}

func (e LabelNewline) GetRoles() []string {
	return e.Roles
}

func (e LabelNewline) GetChildren() []PageInterface {
	return e.Children
}

func (e *LabelNewline) SetChildren(children []PageInterface) {
	e.Children = children
}
