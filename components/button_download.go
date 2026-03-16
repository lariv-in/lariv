package components

import (
	"context"

	"github.com/lariv-in/getters"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

type ButtonDownload struct {
	Page
	Label   string
	Link    getters.Getter[string]
	Classes string
}

func (e ButtonDownload) GetKey() string {
	return e.Key
}

func (e ButtonDownload) GetRoles() []string {
	return e.Roles
}

func (e ButtonDownload) Build(ctx context.Context) Node {
	link := ""
	if e.Link != nil {
		if v, err := e.Link(ctx); err == nil {
			link = v
		}
	}
	return A(
		Href(link),
		Class("btn "+e.Classes),
		Attr("data-hx-boost", "false"),
		Attr("download"),
		Text(e.Label),
	)
}
