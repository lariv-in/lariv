package p_lacerate

import (
	"bytes"
	"context"
	"html"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/url"
	"path"
	"regexp"
	"strings"
	"time"
	"unicode"

	readability "codeberg.org/readeck/go-readability/v2"
	"github.com/JohannesKaufmann/html-to-markdown/v2/converter"
	"github.com/JohannesKaufmann/html-to-markdown/v2/plugin/base"
	"github.com/JohannesKaufmann/html-to-markdown/v2/plugin/commonmark"
	"github.com/JohannesKaufmann/html-to-markdown/v2/plugin/table"
	"github.com/PuerkitoBio/goquery"
)

const (
	maxLinkedArticleBytes  = 4 << 20 // cap download size
	linkedArticleTimeout   = 25 * time.Second
	minLinkedArticleRunes  = 400 // readability / goquery deemed too thin below this
	maxLinkedMarkdownRunes = 56 * 1024
)

var (
	multiNewlineRe      = regexp.MustCompile(`\n{3,}`)
	multiHrBreaksRe     = regexp.MustCompile(`(\n---\n){2,}`)
	linkExtractHTMLConv *converter.Converter
)

func init() {
	linkExtractHTMLConv = converter.NewConverter(
		converter.WithPlugins(
			base.NewBasePlugin(),
			commonmark.NewCommonmarkPlugin(),
			table.NewTablePlugin(),
		),
	)
}

func isSkippableDirectMediaURL(u *url.URL) bool {
	ext := strings.ToLower(path.Ext(u.Path))
	switch ext {
	case ".jpg", ".jpeg", ".png", ".gif", ".webp", ".bmp", ".svg",
		".mp4", ".webm", ".mov", ".mp3", ".wav", ".ogg", ".pdf", ".zip":
		return true
	default:
		return false
	}
}

func isRedditHostedArticleURL(u *url.URL) bool {
	h := strings.ToLower(u.Hostname())
	return h == "reddit.com" || strings.HasSuffix(h, ".reddit.com") ||
		strings.HasSuffix(h, "redd.it")
}

func isPublicIP(ip net.IP) bool {
	if ip == nil {
		return false
	}
	if ip.IsLoopback() || ip.IsPrivate() || ip.IsMulticast() || ip.IsUnspecified() || ip.IsLinkLocalUnicast() {
		return false
	}
	if ip4 := ip.To4(); ip4 != nil {
		if ip4[0] == 169 && ip4[1] == 254 {
			return false
		}
	}
	return true
}

func linkedURLFailsSSRF(ctx context.Context, parsed *url.URL) bool {
	host := strings.TrimSpace(strings.ToLower(parsed.Hostname()))
	if host == "" || host == "localhost" {
		return true
	}
	if ip := net.ParseIP(host); ip != nil {
		return !isPublicIP(ip)
	}
	ips, err := net.DefaultResolver.LookupIPAddr(ctx, host)
	if err != nil || len(ips) == 0 {
		if err != nil {
			slog.Warn("lacerate: linked url dns lookup", "error", err, "host", host)
		}
		return true
	}
	for _, ia := range ips {
		if !isPublicIP(ia.IP) {
			return true
		}
	}
	return false
}

func markdownRuneLen(s string) int {
	return len([]rune(strings.TrimSpace(s)))
}

func markdownTooShort(s string) bool {
	return markdownRuneLen(s) < minLinkedArticleRunes
}

func readabilityHTMLFragment(fullHTML string, pageURL *url.URL) (string, error) {
	article, err := readability.FromReader(strings.NewReader(fullHTML), pageURL)
	if err != nil {
		slog.Warn("lacerate: readability from reader", "error", err, "url", pageURL.Redacted())
		return "", err
	}
	if article.Node == nil {
		return "", nil
	}
	var buf bytes.Buffer
	if err := article.RenderHTML(&buf); err != nil {
		slog.Warn("lacerate: readability render html", "error", err, "url", pageURL.Redacted())
		return "", err
	}
	return strings.TrimSpace(buf.String()), nil
}

func htmlFragmentToMarkdown(htmlFrag, domain string) (string, error) {
	if strings.TrimSpace(htmlFrag) == "" {
		return "", nil
	}
	md, err := linkExtractHTMLConv.ConvertString(htmlFrag, converter.WithDomain(domain))
	if err != nil {
		slog.Warn("lacerate: html to markdown convert", "error", err, "domain", domain)
		return "", err
	}
	return md, nil
}

func markdownFromReadability(fullHTML string, pageURL *url.URL, domain string) string {
	frag, err := readabilityHTMLFragment(fullHTML, pageURL)
	if err != nil {
		return ""
	}
	if frag == "" {
		return ""
	}
	md, err := htmlFragmentToMarkdown(frag, domain)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(md)
}

func goqueryBestFragmentHTML(htmlStr string) string {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlStr))
	if err != nil {
		slog.Warn("lacerate: goquery document", "error", err)
		return ""
	}
	selectors := []string{"article", "[role='main']", "main", "body"}
	best := ""
	bestRunes := 0
	for _, q := range selectors {
		doc.Find(q).Each(func(_ int, s *goquery.Selection) {
			work := s.Clone()
			work.Find("script, style, nav, footer, aside, header, noscript, iframe, form").Remove()
			n := markdownRuneLen(work.Text())
			if n > bestRunes {
				h, err := goquery.OuterHtml(work)
				if err == nil && strings.TrimSpace(h) != "" {
					bestRunes = n
					best = h
				}
			}
		})
	}
	return strings.TrimSpace(best)
}

func markdownFromGoquery(fullHTML, domain string) string {
	frag := goqueryBestFragmentHTML(fullHTML)
	if frag == "" {
		return ""
	}
	md, err := htmlFragmentToMarkdown(frag, domain)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(md)
}

// extractMarkdownFromFetchedHTML tries readability, then goquery, then optional rod-rendered HTML.
func extractMarkdownFromFetchedHTML(ctx context.Context, htmlStr string, pageURL *url.URL, domain string, allowRod bool) string {
	mdR := markdownFromReadability(htmlStr, pageURL, domain)
	if !markdownTooShort(mdR) {
		return mdR
	}
	mdG := markdownFromGoquery(htmlStr, domain)
	if !markdownTooShort(mdG) {
		return mdG
	}
	if allowRod && rodFallbackEnabled() {
		rHTML, err := fetchHTMLViaRod(ctx, pageURL.String())
		if err != nil {
			slog.Warn("lacerate: rod linked article", "error", err, "url", pageURL.Redacted())
		} else if strings.TrimSpace(rHTML) != "" {
			mdRod := extractMarkdownFromFetchedHTML(ctx, rHTML, pageURL, domain, false)
			if !markdownTooShort(mdRod) {
				return mdRod
			}
			if mdRod != "" {
				return mdRod
			}
		}
	}
	if mdG != "" {
		return mdG
	}
	return mdR
}

func normalizeLinkedMarkdown(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	var b strings.Builder
	for line := range strings.SplitSeq(s, "\n") {
		line = strings.TrimRightFunc(line, unicode.IsSpace)
		b.WriteString(line)
		b.WriteByte('\n')
	}
	s = strings.TrimSpace(b.String())
	s = multiNewlineRe.ReplaceAllString(s, "\n\n")
	s = multiHrBreaksRe.ReplaceAllString(s, "\n\n---\n\n")
	rs := []rune(s)
	if len(rs) > maxLinkedMarkdownRunes {
		s = string(rs[:maxLinkedMarkdownRunes]) + "\n\n…[truncated]"
	}
	return s
}

// fetchPostURLAsMarkdown downloads an HTML page linked from a Reddit post, extracts main article content,
// and converts it to markdown. Returns empty string on failure (caller keeps link in metadata only).
func fetchPostURLAsMarkdown(ctx context.Context, raw string) string {
	raw = strings.TrimSpace(html.UnescapeString(raw))
	if raw == "" {
		return ""
	}
	parsed, err := url.Parse(raw)
	if err != nil {
		slog.Warn("lacerate: linked article url parse", "error", err)
		return ""
	}
	if parsed.Host == "" || (parsed.Scheme != "http" && parsed.Scheme != "https") {
		slog.Warn("lacerate: linked article url invalid", "url", raw)
		return ""
	}
	if isSkippableDirectMediaURL(parsed) || isRedditHostedArticleURL(parsed) {
		return ""
	}
	if linkedURLFailsSSRF(ctx, parsed) {
		slog.Warn("lacerate: skip linked fetch (ssrf guard)", "host", parsed.Hostname())
		return ""
	}

	ctx, cancel := context.WithTimeout(ctx, linkedArticleTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, parsed.String(), nil)
	if err != nil {
		slog.Warn("lacerate: linked article new request", "error", err, "url", parsed.Redacted())
		return ""
	}
	req.Header.Set("User-Agent", Config.IntelPreview.UserAgent)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")

	client := &http.Client{Timeout: linkedArticleTimeout}
	resp, err := client.Do(req)
	if err != nil {
		slog.Warn("lacerate: linked article fetch", "error", err, "url", parsed.Redacted())
		return ""
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		slog.Warn("lacerate: linked article status", "status", resp.Status, "url", parsed.Redacted())
		return ""
	}
	ct := strings.ToLower(resp.Header.Get("Content-Type"))
	if !strings.Contains(ct, "text/html") && !strings.Contains(ct, "application/xhtml") {
		return ""
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, maxLinkedArticleBytes))
	if err != nil {
		slog.Warn("lacerate: linked article read body", "error", err, "url", parsed.Redacted())
		return ""
	}
	fullHTML := string(body)
	domain := parsed.Scheme + "://" + parsed.Host

	out := extractMarkdownFromFetchedHTML(ctx, fullHTML, parsed, domain, true)
	out = normalizeLinkedMarkdown(out)
	if out == "" {
		return ""
	}
	return out
}
