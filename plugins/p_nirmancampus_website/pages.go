package p_nirmancampus_website

import (
	"bytes"
	"context"
	"embed"
	"html/template"

	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/lago"
	"maragu.dev/gomponents"
)

//go:embed templates/*.tmpl
var pageTemplatesFS embed.FS

var homePageTmpl = template.Must(template.ParseFS(pageTemplatesFS,
	"templates/home.tmpl",
	"templates/footer.tmpl",
	"templates/header.tmpl",
))

type homeHelloHeading struct {
	components.Page
}

func (e *homeHelloHeading) Build(ctx context.Context) gomponents.Node {
	var buf bytes.Buffer
	if err := homePageTmpl.Execute(&buf, buildHomePageData(ctx)); err != nil {
		panic(err)
	}
	component := gomponents.Raw(buf.String())
	return component
}

func (e *homeHelloHeading) GetKey() string {
	return e.Key
}

func (e *homeHelloHeading) GetRoles() []string {
	return e.Roles
}

func init() {
	lago.RegistryPage.Register("nirmancampus_website.HomePage", &homeHelloHeading{})
}
