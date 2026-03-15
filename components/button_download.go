package components

import (
	"context"
	"fmt"

	"github.com/lariv-in/getters"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

type ButtonDownload struct {
	Page
	Label   string
	Link    getters.Getter
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
		if val := e.Link(ctx); val != nil {
			link = fmt.Sprintf("%s", val)
		}
	}
	return A(
		Href(link),
		Class(fmt.Sprintf("btn %s", e.Classes)),
		Attr("data-hx-boost", "false"),
		Attr("download"),
		Text(e.Label),
	)
}
