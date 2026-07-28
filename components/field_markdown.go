package components

import (
	"context"
	"html/template"
	"io"
	"log/slog"
	"strings"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/ast"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
	"github.com/lariv-in/lariv/getters"
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
	// Sanitize is an optional getter returning whether the rendered markdown HTML should be sanitized.
	Sanitize getters.Getter[bool]
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
	return RenderMarkdownSanitized(md, true, hooks...)
}

// RenderMarkdownSanitized parses and renders a raw markdown string into formatted HTML markup,
// optionally applying HTML sanitization and custom rendering hooks.
func RenderMarkdownSanitized(md string, sanitize bool, hooks ...html.RenderNodeFunc) string {
	doc := ParseMarkdownAST(md)
	opts := html.RendererOptions{Flags: html.CommonFlags}
	if sanitize {
		opts.Flags |= html.SkipHTML | html.NofollowLinks | html.NoopenerLinks | html.NoreferrerLinks
	}
	opts.RenderNodeHook = func(w io.Writer, node ast.Node, entering bool) (ast.WalkStatus, bool) {
		if sanitize && entering {
			if n, ok := node.(*ast.Link); ok {
				dest := strings.TrimSpace(string(n.Destination))
				lower := strings.ToLower(dest)
				if strings.HasPrefix(lower, "javascript:") || strings.HasPrefix(lower, "vbscript:") || strings.HasPrefix(lower, "data:") {
					n.Destination = []byte("#")
				}
			}
		}
		return customRenderHook(w, node, entering)
	}
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

type fieldMarkdownData struct {
	HasContent bool
	Classes    string
	Content    template.HTML
}

// Build compiles the FieldMarkdown component into a Div containing the raw rendered Markdown HTML.
func (e FieldMarkdown) Build(cat Catalog, ctx context.Context, w io.Writer) error {
	if e.Getter == nil {
		return Execute(w, "field_markdown", fieldMarkdownData{})
	}
	s, err := e.Getter(ctx)
	if err != nil {
		slog.Error("FieldMarkdown getter failed", "error", err, "key", e.Key)
		return ContainerError{Error: getters.Static(err)}.Build(cat, ctx, w)
	}
	if s == "" {
		return Execute(w, "field_markdown", fieldMarkdownData{})
	}
	sanitize := true
	if e.Sanitize != nil {
		var err error
		sanitize, err = e.Sanitize(ctx)
		if err != nil {
			slog.Error("FieldMarkdown Sanitize getter failed", "error", err, "key", e.Key)
			return ContainerError{Error: getters.Static(err)}.Build(cat, ctx, w)
		}
	}
	var hooks []html.RenderNodeFunc
	if e.RenderHooks != nil {
		var err error
		hooks, err = e.RenderHooks(ctx, s)
		if err != nil {
			slog.Error("FieldMarkdown RenderHooks failed", "error", err, "key", e.Key)
			return ContainerError{Error: getters.Static(err)}.Build(cat, ctx, w)
		}
	}
	return Execute(w, "field_markdown", fieldMarkdownData{
		HasContent: true,
		Classes:    "whitespace-pre-wrap border border-base-300 p-2 rounded-md " + e.Classes,
		Content:    template.HTML(RenderMarkdownSanitized(s, sanitize, hooks...)),
	})
}
