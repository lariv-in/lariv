package components

import (
	"context"
	"fmt"

	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

type Icon struct {
	Name    string
	Classes string
	Attrs   []Node
}

func (e Icon) Build(ctx context.Context) Node {
	nodes := []Node{
		Class(fmt.Sprintf("heroicon %s", e.Classes)),
		StyleAttr(fmt.Sprintf("--heroicon-url: url('https://api.iconify.design/heroicons/%s.svg')", e.Name)),
	}
	nodes = append(nodes, e.Attrs...)
	return Span(nodes...)
}
