package components

import (
	"context"

	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

// Detail binds an object from context to its descendants under "$in".
// Child fields can then resolve their values via GetterKey("$in.FieldName").
type Detail struct {
	Page
	Getter   Getter
	Children []PageInterface
}

func (e Detail) Build(ctx context.Context) Node {
	value := IfOrGetter(e.Getter, ctx, nil)

	childCtx := ctx
	if value != nil {
		objMap := MapFromStruct(value)
		childCtx = context.WithValue(ctx, "$in", objMap)
	}

	var childNodes []Node
	for _, child := range e.Children {
		childNodes = append(childNodes, Render(child, childCtx))
	}
	return Div(Group(childNodes))
}

func (e Detail) GetChildren() []PageInterface {
	return e.Children
}
