package components

import (
	"context"
	"fmt"

	"maragu.dev/gomponents"
	"maragu.dev/gomponents/html"
)

type Icon struct {
	Page
	Name    string
	Classes string
	Attrs   []gomponents.Node
}

func (e Icon) Build(ctx context.Context) gomponents.Node {
	nodes := []gomponents.Node{
		html.Class(fmt.Sprintf("heroicon %s", e.Classes)),
		html.StyleAttr(fmt.Sprintf("--heroicon-url: url('https://api.iconify.design/heroicons/%s.svg')", e.Name)),
	}
	nodes = append(nodes, e.Attrs...)
	return html.Span(nodes...)
}
