package components

import (
	"context"
	"fmt"

	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

type TopBarButton struct {
	Icon    string
	Url     Getter
	Method  string
	OnClick string
	Classes string
}

type LayoutTopbar struct {
	Buttons  Getter
	Children []PageInterface
}

func (e LayoutTopbar) Build(ctx context.Context) Node {
	var buttons []TopBarButton
	if e.Buttons != nil {
		if val, ok := e.Buttons(ctx).([]TopBarButton); ok {
			buttons = val
		}
	}

	buttonNodes := Group{}
	for _, btn := range buttons {
		url := fmt.Sprintf("%s", IfOrGetter(btn.Url, ctx, ""))
		buttonNodes = append(buttonNodes, Button(
			Class(fmt.Sprintf("btn btn-sm btn-square %s", btn.Classes)),
			If(btn.OnClick != "", Attr("onclick", btn.OnClick)),
			If(url != "", Attr("hx-get", url)),
			If(btn.Method != "", Attr("hx-method", btn.Method)),
			Span(Class(fmt.Sprintf("heroicon heroicon-%s", btn.Icon))),
		))
	}

	childGroup := Group{}
	for _, child := range e.Children {
		childGroup = append(childGroup, child.Build(ctx))
	}

	return Div(Class("h-screen flex flex-col overflow-hidden"),
		Div(Class("navbar bg-base-100 border-b border-base-300 px-4 flex justify-between items-center flex-none"),
			Div(Class("flex-1"),
				A(Href("/"), Class("text-xl font-bold"), Text("Lago")),
			),
			Div(Class("flex-none flex items-center gap-2"),
				buttonNodes,
			),
		),
		Div(Class("flex-1 overflow-hidden"),
			childGroup,
		),
	)
}

func (e LayoutTopbar) GetChildren() []PageInterface {
	return e.Children
}
