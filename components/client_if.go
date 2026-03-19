package components

import (
	"context"

	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

// ClientIf renders children conditionally on the client side using Alpine.
// Condition should be a valid Alpine expression (for example: "isDirectory").
type ClientIf struct {
	Page
	Condition string
	// Optional Alpine data object string, defaults to "{}".
	Data string
	// Optional Alpine init expression.
	Init     string
	Children []PageInterface
}

func (e ClientIf) Build(ctx context.Context) Node {
	data := e.Data
	if data == "" {
		data = "{}"
	}

	group := Group{}
	for _, child := range e.Children {
		group = append(group, Render(child, ctx))
	}

	return Div(
		Attr("x-data", data),
		If(e.Init != "", Attr("x-init", e.Init)),
		If(e.Condition != "", Attr("x-show", e.Condition)),
		group,
	)
}

func (e ClientIf) GetKey() string {
	return e.Key
}

func (e ClientIf) GetRoles() []string {
	return e.Roles
}

func (e ClientIf) GetChildren() []PageInterface {
	return e.Children
}

func (e *ClientIf) SetChildren(children []PageInterface) {
	e.Children = children
}
