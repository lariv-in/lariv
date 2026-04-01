package components

import (
	"context"
	"io"
	"log/slog"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/ast"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
	"github.com/lariv-in/lago/getters"
	"maragu.dev/gomponents"
	ghtml "maragu.dev/gomponents/html"
)

var mdExtensions = parser.CommonExtensions | parser.AutoHeadingIDs

type FieldMarkdown struct {
	Page
	Getter  getters.Getter[string]
	Classes string
}

func appendOrAssign(attr *ast.Attribute, values ...string) *ast.Attribute {
	attribute := attr
	if attr == nil {
		attribute = &ast.Attribute{
			ID:      []byte{},
			Classes: [][]byte{},
			Attrs:   map[string][]byte{},
		}
	}
	for _, v := range values {
		attribute.Classes = append(attribute.Classes, []byte(v))
	}
	return attribute
}

func customRenderHook(w io.Writer, node ast.Node, entering bool) (ast.WalkStatus, bool) {
	if n, ok := node.(*ast.Heading); ok {
		if n.Level == 1 {
			n.Attribute = appendOrAssign(n.Attribute, "text-2xl", "font-bold")
		}
		if n.Level == 2 {
			n.Attribute = appendOrAssign(n.Attribute, "text-xl", "font-semibold")
		}
		if n.Level == 3 {
			n.Attribute = appendOrAssign(n.Attribute, "text-lg", "font-medium")
		}
	}
	if n, ok := node.(*ast.HorizontalRule); ok {
		n.Attribute = appendOrAssign(n.Attribute, "my-4")
	}
	if n, ok := node.(*ast.Paragraph); ok {
		n.Attribute = appendOrAssign(n.Attribute, "my-2")
	}
	if n, ok := node.(*ast.List); ok {
		if n.ListFlags&ast.ListTypeTerm != 0 {
			n.Attribute = appendOrAssign(n.Attribute, "list-disc")
		}
		if n.ListFlags&ast.ListTypeOrdered != 0 {
			n.Attribute = appendOrAssign(n.Attribute, "list-decimal")
		}
		n.Attribute = appendOrAssign(n.Attribute, "my-2", "gap-2", "list-inside")
	}
	return ast.GoToNext, false
}

func RenderMarkdown(md string) string {
	doc := parser.NewWithExtensions(mdExtensions).Parse([]byte(md))
	opts := html.RendererOptions{Flags: html.CommonFlags}
	opts.RenderNodeHook = customRenderHook
	renderer := html.NewRenderer(opts)

	return string(markdown.Render(doc, renderer))
}

func (e FieldMarkdown) GetKey() string {
	return e.Key
}

func (e FieldMarkdown) GetRoles() []string {
	return e.Roles
}

func (e FieldMarkdown) Build(ctx context.Context) gomponents.Node {
	if e.Getter == nil {
		return ghtml.Div()
	}
	s, err := e.Getter(ctx)
	if err != nil {
		slog.Error("FieldMarkdown getter failed", "error", err, "key", e.Key)
		return ContainerError{Error: getters.Static(err)}.Build(ctx)
	}
	if s == "" {
		return ghtml.Div()
	}
	return ghtml.Div(ghtml.Class(e.Classes), gomponents.Raw(RenderMarkdown(s)))
}
