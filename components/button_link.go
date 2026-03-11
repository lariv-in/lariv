package components

import (
	"context"
	"fmt"

	"maragu.dev/gomponents"
	"maragu.dev/gomponents/html"
)

type ButtonLink struct {
	Page
	Label   string
	Link    Getter
	Classes string
}

func (e ButtonLink) Build(ctx context.Context) gomponents.Node {
	link := ""
	if e.Link != nil {
		if val := e.Link(ctx); val != nil {
			link = fmt.Sprintf("%s", val)
		}
	}
	return html.A(html.Href(link), html.Class(fmt.Sprintf("link link-primary %s", e.Classes)), gomponents.Text(e.Label))
}
