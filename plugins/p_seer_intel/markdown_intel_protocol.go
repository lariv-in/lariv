package p_seer_intel

import (
	"bytes"
	"context"
	"fmt"
	"html"
	"io"
	"strconv"
	"strings"

	"github.com/gomarkdown/markdown/ast"
	mdhtml "github.com/gomarkdown/markdown/html"
	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
)

const intelProtoPrefix = "intel://"

// IntelProtocolMarkdownHooks returns render hooks that rewrite [label](intel://<id>) links
// to the Intel detail page href and the row Title as link text. Unknown or missing ids render
// as escaped plain text (no intel:// href).
func IntelProtocolMarkdownHooks(ctx context.Context, md string) ([]mdhtml.RenderNodeFunc, error) {
	db, err := getters.DBFromContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("p_seer_intel: IntelProtocolMarkdownHooks: %w", err)
	}
	doc := components.ParseMarkdownAST(md)
	ids := collectIntelLinkIDs(doc)
	if len(ids) == 0 {
		return nil, nil
	}

	type row struct {
		ID    uint
		Title string
	}
	var rows []row
	if err := db.WithContext(ctx).Model(&Intel{}).Select("id", "title").Where("id IN ?", ids).Find(&rows).Error; err != nil {
		return nil, fmt.Errorf("p_seer_intel: IntelProtocolMarkdownHooks load intel: %w", err)
	}

	titleByID := make(map[uint]string, len(rows))
	hrefByID := make(map[uint]string, len(rows))
	for _, r := range rows {
		titleByID[r.ID] = r.Title
		path, err := lago.RoutePath("seer_intel.DetailRoute", map[string]getters.Getter[any]{
			"id": getters.Any(getters.Static(strconv.FormatUint(uint64(r.ID), 10))),
		})(ctx)
		if err != nil {
			return nil, fmt.Errorf("p_seer_intel: IntelProtocolMarkdownHooks detail path: %w", err)
		}
		hrefByID[r.ID] = path
	}

	hook := func(w io.Writer, node ast.Node, entering bool) (ast.WalkStatus, bool) {
		link, ok := node.(*ast.Link)
		if !ok || !entering {
			return ast.GoToNext, false
		}
		dest := link.Destination
		if !bytes.HasPrefix(dest, []byte(intelProtoPrefix)) {
			return ast.GoToNext, false
		}
		id, ok := parseIntelProtocolID(dest)
		if !ok {
			_, _ = w.Write([]byte(html.EscapeString(linkPlainFallbackText(link))))
			return ast.SkipChildren, true
		}
		href, found := hrefByID[id]
		if !found {
			_, _ = w.Write([]byte(html.EscapeString(linkPlainFallback(link, id))))
			return ast.SkipChildren, true
		}
		title := strings.TrimSpace(titleByID[id])
		if title == "" {
			title = fmt.Sprintf("intel %d", id)
		}
		link.Destination = []byte(href)
		link.Title = nil
		link.Children = []ast.Node{&ast.Text{Leaf: ast.Leaf{Literal: []byte(title)}}}
		return ast.GoToNext, false
	}
	return []mdhtml.RenderNodeFunc{hook}, nil
}

func parseIntelProtocolID(dest []byte) (uint, bool) {
	if !bytes.HasPrefix(dest, []byte(intelProtoPrefix)) {
		return 0, false
	}
	rest := dest[len(intelProtoPrefix):]
	if len(rest) == 0 {
		return 0, false
	}
	for _, b := range rest {
		if b < '0' || b > '9' {
			return 0, false
		}
	}
	n, err := strconv.ParseUint(string(rest), 10, 32)
	if err != nil {
		return 0, false
	}
	return uint(n), true
}

func collectIntelLinkIDs(doc ast.Node) []uint {
	seen := make(map[uint]struct{})
	ast.WalkFunc(doc, func(n ast.Node, entering bool) ast.WalkStatus {
		if !entering {
			return ast.GoToNext
		}
		link, ok := n.(*ast.Link)
		if !ok {
			return ast.GoToNext
		}
		id, ok := parseIntelProtocolID(link.Destination)
		if !ok {
			return ast.GoToNext
		}
		seen[id] = struct{}{}
		return ast.GoToNext
	})
	out := make([]uint, 0, len(seen))
	for id := range seen {
		out = append(out, id)
	}
	return out
}

func appendInlineText(n ast.Node, b *strings.Builder) {
	switch x := n.(type) {
	case *ast.Text:
		b.Write(x.Literal)
	case *ast.Link:
		for _, ch := range x.GetChildren() {
			appendInlineText(ch, b)
		}
	default:
		if c := n.AsContainer(); c != nil {
			for _, ch := range c.Children {
				appendInlineText(ch, b)
			}
		}
	}
}

func linkPlainFallbackText(link *ast.Link) string {
	var buf strings.Builder
	if c := link.AsContainer(); c != nil {
		for _, ch := range c.Children {
			appendInlineText(ch, &buf)
		}
	}
	s := strings.TrimSpace(buf.String())
	if s != "" {
		return s
	}
	return "intel"
}

func linkPlainFallback(link *ast.Link, id uint) string {
	var buf strings.Builder
	if c := link.AsContainer(); c != nil {
		for _, ch := range c.Children {
			appendInlineText(ch, &buf)
		}
	}
	s := strings.TrimSpace(buf.String())
	if s != "" {
		return s
	}
	return fmt.Sprintf("intel %d", id)
}
