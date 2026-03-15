package components

import (
	"context"
	"fmt"

	"github.com/lariv-in/getters"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

// ShowIf renders Children only when Getter resolves to a truthy value.
type ShowIf struct {
	Page
	Getter   getters.Getter
	Children []PageInterface
}

func (e ShowIf) Build(ctx context.Context) Node {
	v := getters.IfOrGetter(e.Getter, ctx, nil)
	if !isTruthy(v) {
		return Group{}
	}
	var nodes []Node
	for _, child := range e.Children {
		nodes = append(nodes, Render(child, ctx))
	}
	return Div(Group(nodes))
}

func (e ShowIf) GetKey() string {
	return e.Key
}

func (e ShowIf) GetRoles() []string {
	return e.Roles
}

func (e ShowIf) GetChildren() []PageInterface {
	return e.Children
}

func (e *ShowIf) SetChildren(children []PageInterface) {
	e.Children = children
}

func isTruthy(v any) bool {
	if v == nil {
		return false
	}
	switch t := v.(type) {
	case bool:
		return t
	case string:
		return t != ""
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		// This is so stupid
		return fmt.Sprintf("%d", t) != "0"
	default:
		return true
	}
}
