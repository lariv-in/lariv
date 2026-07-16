package components

import (
	"context"
	"io"
	"log/slog"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/ast"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
	"github.com/lariv-in/lariv/getters"
	"maragu.dev/gomponents"
	ghtml "maragu.dev/gomponents/html"
)

// MarkdownParserExtensions specifies the common parser extensions used for markdown parsing, including auto heading IDs.
const MarkdownParserExtensions = parser.CommonExtensions | parser.AutoHeadingIDs

// ParseMarkdownAST parses a raw markdown string into a Markdown Abstract Syntax Tree (AST) using [MarkdownParserExtensions].
func ParseMarkdownAST(md string) ast.Node {
	return parser.NewWithExtensions(MarkdownParserExtensions).Parse([]byte(md))
}

// FieldMarkdown represents a read-only field that parses a markdown string and renders it as formatted HTML.
// It formats markdown headers, paragraphs, and list elements to match DaisyUI/Tailwind typography by default.
//
// Use Cases:
//   - Rendering rich-text body content, system descriptions, user posts, or comments stored as markdown.
//
// Example:
//
//	&components.FieldMarkdown{
//	    Getter:  getters.Key[string]("$in.ArticleContent"),
//	    Classes: "prose",
//	}
type FieldMarkdown struct {
	// Page embeds common component properties like Key and Roles.
	Page
	// Getter is the dynamic function retrieving the raw markdown string to render.
	Getter getters.Getter[string]
	// Classes represents additional CSS classes applied to the output HTML div wrapper.
	// (Discouraged: Use layout containers or theme styling instead of custom styling overrides).
	Classes string
	// RenderHooks is an optional function returning custom AST walk hooks.
	// Render hooks run outermost-first before the default styling hooks.
	RenderHooks func(context.Context, string) ([]html.RenderNodeFunc, error)
}

// appendOrAssign is a helper that adds CSS classes to an ast.Attribute object, initializing it if nil.
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

// customRenderHook intercepts markdown nodes (headings, horizontal rules, lists, paragraphs)
// and appends styling class attributes (e.g. text sizes, margins, bullet styles) to render them in standard style.
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

// RenderMarkdown parses and renders a raw markdown string into formatted HTML markup, applying the custom rendering hooks.
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

// GetKey returns the unique key identifier for this FieldMarkdown component.
func (e FieldMarkdown) GetKey() string {
	return e.Key
}

// GetRoles returns the authorized roles required to view this FieldMarkdown.
func (e FieldMarkdown) GetRoles() []string {
	return e.Roles
}

// Build compiles the FieldMarkdown component into a Div Node containing the raw rendered Markdown HTML.
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
