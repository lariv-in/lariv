package p_nirmancampus_website

import (
	"context"

	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/lago"
	"maragu.dev/gomponents"
	"maragu.dev/gomponents/html"
)

type homeHelloHeading struct {
	components.Page
}

func (e *homeHelloHeading) Build(ctx context.Context) gomponents.Node {
	return html.H1(gomponents.Text("hello"))
}

func (e *homeHelloHeading) GetKey() string {
	return e.Key
}

func (e *homeHelloHeading) GetRoles() []string {
	return e.Roles
}

func init() {
	lago.RegistryPage.Register("nirmancampus_website.HomePage", &components.LayoutSimple{
		Children: []components.PageInterface{
			&homeHelloHeading{},
		},
	})
}
