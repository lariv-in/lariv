package p_nirmancampus_website

import (
	"bytes"
	"context"
	"embed"
	"html/template"
	"time"

	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/lago"
	"maragu.dev/gomponents"
)

//go:embed templates/*.tmpl
var pageTemplatesFS embed.FS

var topbarTmpl = template.Must(template.New("topbar.tmpl").Funcs(template.FuncMap{
	"static": websiteStaticPath,
}).ParseFS(
	pageTemplatesFS,
	"templates/topbar.tmpl",
))

var footerTmpl = template.Must(template.New("footer.tmpl").Funcs(template.FuncMap{
	"static": websiteStaticPath,
}).ParseFS(
	pageTemplatesFS,
	"templates/footer.tmpl",
))

var homePageTmpl = template.Must(template.New("home.tmpl").Funcs(template.FuncMap{
	"static": websiteStaticPath,
}).ParseFS(
	pageTemplatesFS,
	"templates/home.tmpl",
))

var contactPageTmpl = template.Must(template.New("contact.tmpl").Funcs(template.FuncMap{
	"static": websiteStaticPath,
}).ParseFS(
	pageTemplatesFS,
	"templates/contact.tmpl",
))

var privacyPageTmpl = template.Must(template.New("privacy.tmpl").Funcs(template.FuncMap{
	"static": websiteStaticPath,
}).ParseFS(
	pageTemplatesFS,
	"templates/privacy.tmpl",
))

var coursesPageTmpl = template.Must(template.New("courses.tmpl").Funcs(template.FuncMap{
	"static": websiteStaticPath,
}).ParseFS(
	pageTemplatesFS,
	"templates/courses.tmpl",
))

var studentZonePageTmpl = template.Must(template.New("student_zone.tmpl").Funcs(template.FuncMap{
	"static": websiteStaticPath,
}).ParseFS(
	pageTemplatesFS,
	"templates/student_zone.tmpl",
))

func isAuthenticated(ctx context.Context) bool {
	return ctx.Value("$user") != nil
}

func renderTopbar(ctx context.Context) string {
	var buf bytes.Buffer
	data := struct{ IsAuthenticated bool }{IsAuthenticated: isAuthenticated(ctx)}
	if err := topbarTmpl.Execute(&buf, data); err != nil {
		panic(err)
	}
	return buf.String()
}

func renderFooter() string {
	var buf bytes.Buffer
	if err := footerTmpl.Execute(&buf, struct{ Year int }{time.Now().Year()}); err != nil {
		panic(err)
	}
	return buf.String()
}

type homePage struct {
	components.Page
}

func (e *homePage) Build(ctx context.Context) gomponents.Node {
	var buf bytes.Buffer
	if err := homePageTmpl.Execute(&buf, buildHomePageData(ctx)); err != nil {
		panic(err)
	}
	return components.Render(components.ShellBase{
		Children: []components.PageInterface{
			&rawPage{content: renderTopbar(ctx)},
			&rawPage{content: buf.String()},
			&rawPage{content: renderFooter()},
		},
	}, ctx)
}

func (e *homePage) GetKey() string   { return e.Key }
func (e *homePage) GetRoles() []string { return e.Roles }

// rawPage wraps raw HTML string as a PageInterface so it can be a ShellBase child.
type rawPage struct {
	components.Page
	content string
}

func (r *rawPage) Build(_ context.Context) gomponents.Node {
	return gomponents.Raw(r.content)
}

func (r *rawPage) GetKey() string     { return r.Key }
func (r *rawPage) GetRoles() []string { return r.Roles }

type coursesPage struct {
	components.Page
}

func (e *coursesPage) Build(ctx context.Context) gomponents.Node {
	var buf bytes.Buffer
	if err := coursesPageTmpl.Execute(&buf, buildCoursesPageData(ctx)); err != nil {
		panic(err)
	}
	return components.Render(components.ShellBase{
		Children: []components.PageInterface{
			&rawPage{content: renderTopbar(ctx)},
			&rawPage{content: buf.String()},
			&rawPage{content: renderFooter()},
		},
	}, ctx)
}

func (e *coursesPage) GetKey() string     { return e.Key }
func (e *coursesPage) GetRoles() []string { return e.Roles }

type contactPage struct {
	components.Page
}

func (e *contactPage) Build(ctx context.Context) gomponents.Node {
	var buf bytes.Buffer
	if err := contactPageTmpl.Execute(&buf, struct{}{}); err != nil {
		panic(err)
	}
	return components.Render(components.ShellBase{
		Children: []components.PageInterface{
			&rawPage{content: renderTopbar(ctx)},
			&rawPage{content: buf.String()},
			&rawPage{content: renderFooter()},
		},
	}, ctx)
}

func (e *contactPage) GetKey() string     { return e.Key }
func (e *contactPage) GetRoles() []string { return e.Roles }

type privacyPage struct {
	components.Page
}

func (e *privacyPage) Build(ctx context.Context) gomponents.Node {
	var buf bytes.Buffer
	if err := privacyPageTmpl.Execute(&buf, struct{}{}); err != nil {
		panic(err)
	}
	return components.Render(components.ShellBase{
		Children: []components.PageInterface{
			&rawPage{content: renderTopbar(ctx)},
			&rawPage{content: buf.String()},
			&rawPage{content: renderFooter()},
		},
	}, ctx)
}

func (e *privacyPage) GetKey() string     { return e.Key }
func (e *privacyPage) GetRoles() []string { return e.Roles }

type studentZonePage struct {
	components.Page
}

func (e *studentZonePage) Build(ctx context.Context) gomponents.Node {
	var buf bytes.Buffer
	if err := studentZonePageTmpl.Execute(&buf, buildStudentZonePageData(ctx)); err != nil {
		panic(err)
	}
	return components.Render(components.ShellBase{
		Children: []components.PageInterface{
			&rawPage{content: renderTopbar(ctx)},
			&rawPage{content: buf.String()},
			&rawPage{content: renderFooter()},
		},
	}, ctx)
}

func (e *studentZonePage) GetKey() string     { return e.Key }
func (e *studentZonePage) GetRoles() []string { return e.Roles }

func init() {
	lago.RegistryPage.Register("nirmancampus_website.HomePage", &homePage{})
	lago.RegistryPage.Register("nirmancampus_website.CoursesPage", &coursesPage{})
	lago.RegistryPage.Register("nirmancampus_website.ContactPage", &contactPage{})
	lago.RegistryPage.Register("nirmancampus_website.PrivacyPage", &privacyPage{})
	lago.RegistryPage.Register("nirmancampus_website.StudentZonePage", &studentZonePage{})
}
