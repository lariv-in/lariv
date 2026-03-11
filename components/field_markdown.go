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

var mdParser = parser.NewWithExtensions(parser.CommonExtensions | parser.AutoHeadingIDs )

type FieldMarkdown struct {
	Page
	Getter  getters.Getter
	Classes string
}

func (e FieldMarkdown) Build(ctx context.Context) Node {
	s, _ := getters.IfOrGetter(e.Getter, ctx, "").(string)
	if s == "" {
		return ghtml.Div()
	}
	doc := mdParser.Parse([]byte(s))
	opts := html.RendererOptions{Flags: html.CommonFlags}
	renderer := html.NewRenderer(opts)
	rendered := string(markdown.Render(doc, renderer))
	return ghtml.Div(ghtml.Class(e.Classes), Raw(rendered))
}
