package components

import (
	"context"

	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

// ClientShow wraps children in a div with x-show only (no x-data). Use inside a parent
// Alpine scope (e.g. ClientData) so the condition can read parent state.
type ClientShow struct {
	Page
	Condition string
	Children  []PageInterface
}

func (e ClientShow) Build(ctx context.Context) Node {
	group := Group{}
	for _, child := range e.Children {
		group = append(group, Render(child, ctx))
	}
	return Div(
		If(e.Condition != "", Attr("x-show", e.Condition)),
		group,
	)
}

func (e ClientShow) GetKey() string {
	return e.Key
}

func (e ClientShow) GetRoles() []string {
	return e.Roles
}

func (e ClientShow) GetChildren() []PageInterface {
	return e.Children
}

func (e *ClientShow) SetChildren(children []PageInterface) {
	e.Children = children
}
