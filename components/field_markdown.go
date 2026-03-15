package components

import (
	"context"
	"strings"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
	"github.com/lariv-in/getters"
	. "maragu.dev/gomponents"
	ghtml "maragu.dev/gomponents/html"
)

var mdExtensions = parser.CommonExtensions | parser.AutoHeadingIDs

type FieldMarkdown struct {
	Page
	Getter  getters.Getter
	Classes string
}

func RenderMarkdown(md string) string {
	doc := parser.NewWithExtensions(mdExtensions).Parse([]byte(md))
	opts := html.RendererOptions{Flags: html.CommonFlags}
	renderer := html.NewRenderer(opts)
	s := string(markdown.Render(doc, renderer))

	// Add Tailwind-style classes to headings
	s = strings.ReplaceAll(s, `<h1 id="`, `<h1 class="text-2xl font-bold" id="`)
	s = strings.ReplaceAll(s, `<h2 id="`, `<h2 class="text-xl font-semibold" id="`)
	s = strings.ReplaceAll(s, `<h3 id="`, `<h3 class="text-lg font-medium" id="`)

	// Add vertical margin to horizontal rules
	s = strings.ReplaceAll(s, "<hr", `<hr class="my-4"`)
	s = strings.ReplaceAll(s, "<p", `<p class="my-2"`)
	s = strings.ReplaceAll(s, "<ul", `<ul class="list-disc m-2 gap-1"`)
	s = strings.ReplaceAll(s, "<ol", `<ol class="list-decimal m-2 gap-1"`)

	return s
}

func (e FieldMarkdown) GetKey() string {
	return e.Key
}

func (e FieldMarkdown) GetRoles() []string {
	return e.Roles
}

func (e FieldMarkdown) Build(ctx context.Context) Node {
	s, _ := getters.IfOrGetter(e.Getter, ctx, "").(string)
	if s == "" {
		return ghtml.Div()
	}
	return ghtml.Div(ghtml.Class(e.Classes), Raw(RenderMarkdown(s)))
}
