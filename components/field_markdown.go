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

// MarkdownParserExtensions matches [FieldMarkdown] / [RenderMarkdown] parsing.
const MarkdownParserExtensions = parser.CommonExtensions | parser.AutoHeadingIDs

// ParseMarkdownAST parses markdown with the same extensions as [RenderMarkdown].
func ParseMarkdownAST(md string) ast.Node {
	return parser.NewWithExtensions(MarkdownParserExtensions).Parse([]byte(md))
}

type FieldMarkdown struct {
	Page
	Getter  getters.Getter[string]
	Classes string
	// RenderHooks is optional. When non-nil, called with request context and the markdown
	// string from Getter; returned hooks run outermost-first (before built-in styling hooks).
	RenderHooks func(context.Context, string) ([]html.RenderNodeFunc, error)
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
	if !entering {
		return ast.GoToNext, false
	}
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
		// ListTypeTerm is definition-list markup, not bullet UL. Bullet / loose lists
		// have neither Ordered nor Term; they still need list-disc for Tailwind.
		if n.ListFlags&ast.ListTypeOrdered != 0 {
			n.Attribute = appendOrAssign(n.Attribute, "list-decimal")
		} else {
			n.Attribute = appendOrAssign(n.Attribute, "list-disc")
		}
		n.Attribute = appendOrAssign(n.Attribute, "my-2", "gap-2", "list-inside")
	}
	return ast.GoToNext, false
}

func RenderMarkdown(md string, hooks ...html.RenderNodeFunc) string {
	doc := ParseMarkdownAST(md)
	opts := html.RendererOptions{Flags: html.CommonFlags}
	opts.RenderNodeHook = customRenderHook
	for _, renderNodeFunc := range hooks {
		currentFunc := opts.RenderNodeHook
		opts.RenderNodeHook = func(w io.Writer, node ast.Node, entering bool) (ast.WalkStatus, bool) {
			status, processed := renderNodeFunc(w, node, entering)
			if !processed {
				return currentFunc(w, node, entering)
			}
			return status, processed
		}
	}
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
	var hooks []html.RenderNodeFunc
	if e.RenderHooks != nil {
		var err error
		hooks, err = e.RenderHooks(ctx, s)
		if err != nil {
			slog.Error("FieldMarkdown RenderHooks failed", "error", err, "key", e.Key)
			return ContainerError{Error: getters.Static(err)}.Build(ctx)
		}
	}
	return ghtml.Div(
		ghtml.Class("whitespace-pre-wrap border border-base-300 p-2 rounded-md "+e.Classes),
		gomponents.Raw(RenderMarkdown(s, hooks...)),
	)
}
