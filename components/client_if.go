package components

import (
	"context"

	. "maragu.dev/gomponents"
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
	group := Group{}
	for _, child := range e.Children {
		group = append(group, Render(child, ctx))
	}

	content := Node(group)
	if len(group) > 1 {
		// Alpine x-if template content must have a single root element.
		content = El("div", group)
	}

	return El("template",
		If(e.Condition != "", Attr("x-if", e.Condition)),
		content,
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
