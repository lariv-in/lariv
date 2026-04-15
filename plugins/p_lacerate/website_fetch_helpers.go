package p_lacerate

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"html"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

const websiteMaxChildLinks = 12

type websiteFetchedPage struct {
	PageURL *url.URL
	HTML    string
	Doc     *goquery.Document
}

func normalizeWebsiteSeedURL(raw string) (string, error) {
	return normalizeWebsiteResolvedURL(nil, raw)
}

func normalizeWebsiteResolvedURL(base *url.URL, raw string) (string, error) {
	raw = strings.TrimSpace(html.UnescapeString(raw))
	if raw == "" {
		return "", fmt.Errorf("url is required")
	}
	parsed, err := url.Parse(raw)
	if err != nil {
		return "", fmt.Errorf("invalid url %q: %w", raw, err)
	}
	if base != nil {
		parsed = base.ResolveReference(parsed)
	}
	return normalizeWebsiteParsedURL(parsed)
}

func normalizeWebsiteParsedURL(parsed *url.URL) (string, error) {
	if parsed == nil {
		return "", fmt.Errorf("url is required")
	}
	if parsed.Host == "" || parsed.Scheme == "" {
		return "", fmt.Errorf("url must be absolute http(s)")
	}
	out := *parsed
	out.User = nil
	out.Scheme = strings.ToLower(out.Scheme)
	switch out.Scheme {
	case "http", "https":
	default:
		return "", fmt.Errorf("url must use http or https")
	}
	host := strings.ToLower(out.Hostname())
	if host == "" {
		return "", fmt.Errorf("url host is required")
	}
	port := out.Port()
	switch {
	case port == "":
		out.Host = host
	case out.Scheme == "http" && port == "80":
		out.Host = host
	case out.Scheme == "https" && port == "443":
		out.Host = host
	default:
		out.Host = net.JoinHostPort(host, port)
	}
	out.Fragment = ""
	if out.Path == "" {
		out.Path = "/"
	}
	query := out.Query()
	out.RawQuery = query.Encode()
	return out.String(), nil
}

func websiteFetchableURL(ctx context.Context, raw string) (*url.URL, string, error) {
	normalized, err := normalizeWebsiteSeedURL(raw)
	if err != nil {
		return nil, "", err
	}
	parsed, err := url.Parse(normalized)
	if err != nil {
		return nil, "", err
	}
	if isSkippableDirectMediaURL(parsed) {
		return nil, "", fmt.Errorf("url points to direct media: %s", normalized)
	}
	if linkedURLFailsSSRF(ctx, parsed) {
		return nil, "", fmt.Errorf("url blocked by ssrf guard: %s", normalized)
	}
	return parsed, normalized, nil
}

func fetchWebsitePage(ctx context.Context, raw string) (websiteFetchedPage, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	_, normalized, err := websiteFetchableURL(ctx, raw)
	if err != nil {
		return websiteFetchedPage{}, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, normalized, nil)
	if err != nil {
		return websiteFetchedPage{}, err
	}
	req.Header.Set("User-Agent", Config.IntelPreview.UserAgent)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	client := &http.Client{
		Timeout: linkedArticleTimeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 10 {
				return fmt.Errorf("too many redirects for %s", normalized)
			}
			if linkedURLFailsSSRF(ctx, req.URL) {
				return fmt.Errorf("redirect blocked by ssrf guard: %s", req.URL.String())
			}
			return nil
		},
	}
	resp, err := client.Do(req)
	if err != nil {
		return websiteFetchedPage{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return websiteFetchedPage{}, fmt.Errorf("website fetch status %s", resp.Status)
	}
	ct := strings.ToLower(resp.Header.Get("Content-Type"))
	if !strings.Contains(ct, "text/html") && !strings.Contains(ct, "application/xhtml") {
		return websiteFetchedPage{}, fmt.Errorf("website fetch content-type %q is not html", ct)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, maxLinkedArticleBytes))
	if err != nil {
		return websiteFetchedPage{}, err
	}
	finalURL, err := normalizeWebsiteParsedURL(resp.Request.URL)
	if err != nil {
		return websiteFetchedPage{}, err
	}
	parsedFinalURL, err := url.Parse(finalURL)
	if err != nil {
		return websiteFetchedPage{}, err
	}
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(body)))
	if err != nil {
		return websiteFetchedPage{}, err
	}
	return websiteFetchedPage{
		PageURL: parsedFinalURL,
		HTML:    string(body),
		Doc:     doc,
	}, nil
}

func websiteFirstAttr(doc *goquery.Document, selector, attr string) string {
	if doc == nil {
		return ""
	}
	v, _ := doc.Find(selector).First().Attr(attr)
	return strings.TrimSpace(v)
}

func websiteCanonicalURL(doc *goquery.Document, pageURL *url.URL) string {
	candidates := []string{
		websiteFirstAttr(doc, `link[rel="canonical"]`, "href"),
		websiteFirstAttr(doc, `meta[property="og:url"]`, "content"),
		websiteFirstAttr(doc, `meta[name="twitter:url"]`, "content"),
	}
	for _, candidate := range candidates {
		if candidate == "" {
			continue
		}
		if normalized, err := normalizeWebsiteResolvedURL(pageURL, candidate); err == nil {
			return normalized
		}
	}
	if pageURL == nil {
		return ""
	}
	normalized, err := normalizeWebsiteParsedURL(pageURL)
	if err != nil {
		return ""
	}
	return normalized
}

func websitePageTitle(doc *goquery.Document) string {
	if doc == nil {
		return ""
	}
	candidates := []string{
		websiteFirstAttr(doc, `meta[property="og:title"]`, "content"),
		websiteFirstAttr(doc, `meta[name="twitter:title"]`, "content"),
		strings.TrimSpace(doc.Find("title").First().Text()),
		strings.TrimSpace(doc.Find("h1").First().Text()),
	}
	for _, candidate := range candidates {
		if candidate != "" {
			return candidate
		}
	}
	return ""
}

func parseWebsiteTime(raw string) time.Time {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return time.Time{}
	}
	formats := []string{
		time.RFC3339Nano,
		time.RFC3339,
		"2006-01-02T15:04:05.000Z07:00",
		"2006-01-02T15:04:05Z07:00",
		"2006-01-02 15:04:05Z07:00",
		"2006-01-02",
	}
	for _, format := range formats {
		if t, err := time.Parse(format, raw); err == nil {
			return t.UTC()
		}
	}
	return time.Time{}
}

func websitePublishedTime(doc *goquery.Document) time.Time {
	if doc == nil {
		return time.Time{}
	}
	candidates := []string{
		websiteFirstAttr(doc, `meta[property="article:published_time"]`, "content"),
		websiteFirstAttr(doc, `meta[name="parsely-pub-date"]`, "content"),
		websiteFirstAttr(doc, `meta[itemprop="datePublished"]`, "content"),
		websiteFirstAttr(doc, `meta[property="og:updated_time"]`, "content"),
		websiteFirstAttr(doc, `time[datetime]`, "datetime"),
	}
	for _, candidate := range candidates {
		if t := parseWebsiteTime(candidate); !t.IsZero() {
			return t
		}
	}
	return time.Time{}
}

func websiteLinksFromSelection(ctx context.Context, sel *goquery.Selection, pageURL *url.URL) []string {
	if sel == nil || pageURL == nil {
		return nil
	}
	seen := map[string]struct{}{}
	out := make([]string, 0, websiteMaxChildLinks)
	sel.Find("a[href]").EachWithBreak(func(_ int, s *goquery.Selection) bool {
		href, ok := s.Attr("href")
		if !ok {
			return true
		}
		href = strings.TrimSpace(href)
		if href == "" ||
			strings.HasPrefix(href, "#") ||
			strings.HasPrefix(strings.ToLower(href), "mailto:") ||
			strings.HasPrefix(strings.ToLower(href), "javascript:") {
			return true
		}
		resolved, err := normalizeWebsiteResolvedURL(pageURL, href)
		if err != nil {
			return true
		}
		if _, _, err := websiteFetchableURL(ctx, resolved); err != nil {
			slog.Warn("lacerate: skip child website url", "error", err, "url", resolved)
			return true
		}
		if _, ok := seen[resolved]; ok {
			return true
		}
		seen[resolved] = struct{}{}
		out = append(out, resolved)
		return len(out) < websiteMaxChildLinks
	})
	return out
}

func websiteSelectionMarkdown(sel *goquery.Selection, pageURL *url.URL) string {
	if sel == nil || sel.Length() == 0 || pageURL == nil {
		return ""
	}
	work := sel.Clone()
	work.Find("script, style, nav, footer, aside, header, noscript, iframe, form").Remove()
	htmlFrag, err := goquery.OuterHtml(work)
	if err != nil {
		return ""
	}
	domain := pageURL.Scheme + "://" + pageURL.Host
	md, err := htmlFragmentToMarkdown(htmlFrag, domain)
	if err != nil {
		return ""
	}
	return normalizeLinkedMarkdown(md)
}

func websiteBuildContent(title, body, canonicalURL string) string {
	title = strings.TrimSpace(title)
	body = normalizeLinkedMarkdown(body)
	var b strings.Builder
	if title != "" && !strings.HasPrefix(body, "#") {
		b.WriteString("# ")
		b.WriteString(title)
		b.WriteString("\n\n")
	}
	if body != "" {
		b.WriteString(body)
		b.WriteString("\n\n")
	}
	if canonicalURL != "" {
		b.WriteString("---\n\n")
		fmt.Fprintf(&b, "- **Link:** %s\n", canonicalURL)
	}
	return strings.TrimSpace(b.String())
}

func websiteDedupHash(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}

func websiteIntelFromPage(title, body, canonicalURL string, published, fallback time.Time) (Intel, error) {
	content := websiteBuildContent(title, body, canonicalURL)
	if content == "" {
		return Intel{}, fmt.Errorf("website content empty for %s", canonicalURL)
	}
	dedup := websiteDedupHash(canonicalURL)
	if dedup == "" {
		return Intel{}, fmt.Errorf("website dedup hash empty for %s", canonicalURL)
	}
	if published.IsZero() {
		published = fallback
	}
	if published.IsZero() {
		published = time.Now().UTC()
	}
	return Intel{
		DedupHash: &dedup,
		Content:   content,
		Datetime:  published.UTC(),
	}, nil
}
