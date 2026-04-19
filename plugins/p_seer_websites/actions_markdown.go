package p_seer_websites

import (
	"bytes"
	"log/slog"
	"net/url"
	"strings"

	readability "codeberg.org/readeck/go-readability/v2"
	"github.com/JohannesKaufmann/html-to-markdown/v2/converter"
	"github.com/JohannesKaufmann/html-to-markdown/v2/plugin/base"
	"github.com/JohannesKaufmann/html-to-markdown/v2/plugin/commonmark"
	"github.com/JohannesKaufmann/html-to-markdown/v2/plugin/table"
)

const maxScrapedHTMLBytes = 4 << 20

var htmlToMDConverter *converter.Converter

func init() {
	htmlToMDConverter = converter.NewConverter(
		converter.WithPlugins(
			base.NewBasePlugin(),
			commonmark.NewCommonmarkPlugin(),
			table.NewTablePlugin(),
		),
	)
}

func readabilityHTMLFragment(fullHTML string, pageURL *url.URL) (string, error) {
	article, err := readability.FromReader(strings.NewReader(fullHTML), pageURL)
	if err != nil {
		return "", err
	}
	if article.Node == nil {
		return "", nil
	}
	var buf bytes.Buffer
	if err := article.RenderHTML(&buf); err != nil {
		return "", err
	}
	return strings.TrimSpace(buf.String()), nil
}

func htmlFragmentToMarkdown(htmlFrag, domain string) (string, error) {
	if strings.TrimSpace(htmlFrag) == "" {
		return "", nil
	}
	return htmlToMDConverter.ConvertString(htmlFrag, converter.WithDomain(domain))
}

// markdownFromRenderedHTML runs readability then HTML→markdown.
func markdownFromRenderedHTML(fullHTML string, pageURL *url.URL) string {
	domain := ""
	if pageURL != nil {
		domain = pageURL.Hostname()
	}
	frag, err := readabilityHTMLFragment(fullHTML, pageURL)
	if err != nil {
		slog.Warn("p_seer_websites: readability", "error", err, "host", domain)
		return ""
	}
	if frag == "" {
		return ""
	}
	md, err := htmlFragmentToMarkdown(frag, domain)
	if err != nil {
		slog.Warn("p_seer_websites: html to markdown", "error", err, "host", domain)
		return ""
	}
	return strings.TrimSpace(md)
}
