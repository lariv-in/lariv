package components

import (
	"context"

	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

type AccordionItem struct {
	Page
	Title    PageInterface
	Open     bool
	Children []PageInterface
}

type Accordion struct {
	Page
	Classes string
	Items   []AccordionItem
}

func (e Accordion) Build(ctx context.Context) Node {
	var nodes []Node
	for _, item := range e.Items {
		var childNodes []Node
		for _, child := range item.Children {
			childNodes = append(childNodes, Render(child, ctx))
		}
		nodes = append(nodes,
			Div(Class("collapse collapse-arrow bg-base-100 border border-base-300"),
				El("input", Type("checkbox"), If(item.Open, Attr("checked", "checked"))),
				Div(Class("collapse-title"), Render(item.Title, ctx)),
				Div(Class("collapse-content"), Group(childNodes)),
			),
		)
	}
	return Div(Class("join join-vertical w-full "+e.Classes), Group(nodes))
}

func (e Accordion) GetChildren() []PageInterface {
	var all []PageInterface
	for _, item := range e.Items {
		all = append(all, item.Children...)
	}
	return all
}
