package p_seer_websites

import (
	"net/url"
	"strings"
	"testing"
)

func TestMarkdownFromRenderedHTML(t *testing.T) {
	html := `<!doctype html><html><head><title>Hi</title></head><body><article><p>Hello world paragraph one.</p><p>Paragraph two with more words for readability length.</p></article></body></html>`
	u, _ := url.Parse("https://example.com/page")
	md := markdownFromRenderedHTML(html, u)
	if !strings.Contains(md, "Hello world") {
		t.Fatalf("expected readable text in markdown, got %q", md)
	}
}
