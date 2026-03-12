package components

import (
	"context"

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
	return string(markdown.Render(doc, renderer))
}

func (e FieldMarkdown) Build(ctx context.Context) Node {
	s, _ := getters.IfOrGetter(e.Getter, ctx, "").(string)
	if s == "" {
		return ghtml.Div()
	}
	return ghtml.Div(ghtml.Class(e.Classes), Raw(RenderMarkdown(s)))
}
