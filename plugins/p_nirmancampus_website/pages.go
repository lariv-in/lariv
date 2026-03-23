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

var homePageTmpl = template.Must(template.New("home.tmpl").Funcs(template.FuncMap{
	"static": websiteStaticPath,
}).ParseFS(
	pageTemplatesFS,
	"templates/home.tmpl",
	"templates/footer.tmpl",
	"templates/header.tmpl",
))

var coursesPageTmpl = template.Must(template.New("courses.tmpl").Funcs(template.FuncMap{
	"static": websiteStaticPath,
}).ParseFS(
	pageTemplatesFS,
	"templates/courses.tmpl",
	"templates/footer.tmpl",
	"templates/header.tmpl",
))

var aboutUsPageTmpl = template.Must(template.New("about_us.tmpl").Funcs(template.FuncMap{
	"static": websiteStaticPath,
}).ParseFS(
	pageTemplatesFS,
	"templates/about_us.tmpl",
	"templates/footer.tmpl",
	"templates/header.tmpl",
))

var privacyPolicyPageTmpl = template.Must(template.New("privacy_policy.tmpl").Funcs(template.FuncMap{
	"static": websiteStaticPath,
}).ParseFS(
	pageTemplatesFS,
	"templates/privacy_policy.tmpl",
	"templates/footer.tmpl",
	"templates/header.tmpl",
))

var mrscmtPageTmpl = template.Must(template.New("mrscmt.tmpl").Funcs(template.FuncMap{
	"static": websiteStaticPath,
}).ParseFS(
	pageTemplatesFS,
	"templates/mrscmt.tmpl",
	"templates/footer.tmpl",
	"templates/header.tmpl",
))

var mrsptuadmcoPageTmpl = template.Must(template.New("mrsptuadmco.tmpl").Funcs(template.FuncMap{
	"static": websiteStaticPath,
}).ParseFS(
	pageTemplatesFS,
	"templates/mrsptuadmco.tmpl",
	"templates/footer.tmpl",
	"templates/header.tmpl",
))

var oss2281PageTmpl = template.Must(template.New("oss2281.tmpl").Funcs(template.FuncMap{
	"static": websiteStaticPath,
}).ParseFS(
	pageTemplatesFS,
	"templates/oss2281.tmpl",
	"templates/footer.tmpl",
	"templates/header.tmpl",
))

type homeHelloHeading struct {
	components.Page
}

type coursesOfferedPage struct {
	components.Page
}

type aboutUsPage struct {
	components.Page
}

type privacyPolicyPage struct {
	components.Page
}

type mrscmtPage struct {
	components.Page
}

type mrsptuadmcoPage struct {
	components.Page
}

type oss2281Page struct {
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

func (e *coursesOfferedPage) Build(ctx context.Context) gomponents.Node {
	var buf bytes.Buffer
	if err := coursesPageTmpl.Execute(&buf, buildCoursesPageData(ctx)); err != nil {
		panic(err)
	}
	component := gomponents.Raw(buf.String())
	return component
}

func (e *aboutUsPage) Build(ctx context.Context) gomponents.Node {
	var buf bytes.Buffer
	if err := aboutUsPageTmpl.Execute(&buf, struct{}{}); err != nil {
		panic(err)
	}
	component := gomponents.Raw(buf.String())
	return component
}

func (e *privacyPolicyPage) Build(ctx context.Context) gomponents.Node {
	var buf bytes.Buffer
	if err := privacyPolicyPageTmpl.Execute(&buf, struct{}{}); err != nil {
		panic(err)
	}
	component := gomponents.Raw(buf.String())
	return component
}

func (e *mrscmtPage) Build(ctx context.Context) gomponents.Node {
	var buf bytes.Buffer
	if err := mrscmtPageTmpl.Execute(&buf, struct{}{}); err != nil {
		panic(err)
	}
	component := gomponents.Raw(buf.String())
	return component
}

func (e *mrsptuadmcoPage) Build(ctx context.Context) gomponents.Node {
	var buf bytes.Buffer
	if err := mrsptuadmcoPageTmpl.Execute(&buf, struct{}{}); err != nil {
		panic(err)
	}
	component := gomponents.Raw(buf.String())
	return component
}

func (e *oss2281Page) Build(ctx context.Context) gomponents.Node {
	var buf bytes.Buffer
	if err := oss2281PageTmpl.Execute(&buf, struct{}{}); err != nil {
		panic(err)
	}
	component := gomponents.Raw(buf.String())
	return component
}

func (e *homeHelloHeading) GetKey() string {
	return e.Key
}

func (e *coursesOfferedPage) GetKey() string {
	return e.Key
}

func (e *aboutUsPage) GetKey() string {
	return e.Key
}

func (e *privacyPolicyPage) GetKey() string {
	return e.Key
}

func (e *mrscmtPage) GetKey() string {
	return e.Key
}

func (e *mrsptuadmcoPage) GetKey() string {
	return e.Key
}

func (e *oss2281Page) GetKey() string {
	return e.Key
}

func (e *homeHelloHeading) GetRoles() []string {
	return e.Roles
}

func (e *coursesOfferedPage) GetRoles() []string {
	return e.Roles
}

func (e *aboutUsPage) GetRoles() []string {
	return e.Roles
}

func (e *privacyPolicyPage) GetRoles() []string {
	return e.Roles
}

func (e *mrscmtPage) GetRoles() []string {
	return e.Roles
}

func (e *mrsptuadmcoPage) GetRoles() []string {
	return e.Roles
}

func (e *oss2281Page) GetRoles() []string {
	return e.Roles
}

func init() {
	lago.RegistryPage.Register("nirmancampus_website.HomePage", &homeHelloHeading{})
	lago.RegistryPage.Register("nirmancampus_website.CoursesPage", &coursesOfferedPage{})
	lago.RegistryPage.Register("nirmancampus_website.AboutUsPage", &aboutUsPage{})
	lago.RegistryPage.Register("nirmancampus_website.PrivacyPolicyPage", &privacyPolicyPage{})
	lago.RegistryPage.Register("nirmancampus_website.MrscmtPage", &mrscmtPage{})
	lago.RegistryPage.Register("nirmancampus_website.MrsptuadmcoPage", &mrsptuadmcoPage{})
	lago.RegistryPage.Register("nirmancampus_website.Oss2281Page", &oss2281Page{})
}
