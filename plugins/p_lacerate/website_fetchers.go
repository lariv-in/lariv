package p_lacerate

import (
	"context"
	"fmt"
	"log/slog"
	"net/url"
	"sort"
	"strings"
	"time"

	"gorm.io/gorm"
)

type WebsiteFetcher interface {
	Priority(url string) int
	FetchWebsite(url string, depth int) ([]Intel, error)
	Match(url string) bool
}

// Named slice type, not alias, so Go can attach methods.
type WebsiteFetchers []WebsiteFetcher

type websiteFetchState struct {
	ctx      context.Context
	db       *gorm.DB
	now      time.Time
	seen     map[string]struct{}
	fetchers WebsiteFetchers
}

type GeneralWebsiteFetcher struct {
	state *websiteFetchState
}

type MediumWebsiteScraper struct {
	state *websiteFetchState
}

var WebsiteFetcherImplementors = WebsiteFetchers{
	&MediumWebsiteScraper{},
	&GeneralWebsiteFetcher{},
}

func NewWebsiteFetchers(ctx context.Context, db *gorm.DB) WebsiteFetchers {
	if ctx == nil {
		ctx = context.Background()
	}
	state := &websiteFetchState{
		ctx:  ctx,
		db:   db,
		now:  time.Now().UTC(),
		seen: map[string]struct{}{},
	}
	out := make(WebsiteFetchers, 0, len(WebsiteFetcherImplementors))
	for _, fetcher := range WebsiteFetcherImplementors {
		switch fetcher.(type) {
		case *MediumWebsiteScraper:
			out = append(out, &MediumWebsiteScraper{state: state})
		case *GeneralWebsiteFetcher:
			out = append(out, &GeneralWebsiteFetcher{state: state})
		default:
			panic(fmt.Sprintf("unsupported website fetcher implementor %T", fetcher))
		}
	}
	state.fetchers = out
	return out
}

func (f WebsiteFetchers) FetchWebsite(url string, depth int) ([]Intel, error) {
	if depth <= 0 {
		return nil, fmt.Errorf("website fetch depth must be >= 1")
	}
	matched := make([]WebsiteFetcher, 0, len(f))
	for _, fetcher := range f {
		if fetcher.Match(url) {
			matched = append(matched, fetcher)
		}
	}
	if len(matched) == 0 {
		return nil, fmt.Errorf("no website fetcher matched %q", url)
	}
	sort.SliceStable(matched, func(i, j int) bool {
		return matched[i].Priority(url) > matched[j].Priority(url)
	})
	var lastErr error
	for _, fetcher := range matched {
		intels, err := fetcher.FetchWebsite(url, depth)
		if err != nil {
			lastErr = err
			slog.Error("lacerate: website fetcher failed", "fetcher", fmt.Sprintf("%T", fetcher), "url", url, "depth", depth, "error", err)
			continue
		}
		return dedupWebsiteIntels(intels), nil
	}
	if lastErr == nil {
		lastErr = fmt.Errorf("all website fetchers failed for %q", url)
	} else {
		lastErr = fmt.Errorf("all website fetchers failed for %q: %w", url, lastErr)
	}
	return nil, lastErr
}

func (s *websiteFetchState) markSeen(normalizedURL string) bool {
	if normalizedURL == "" {
		return false
	}
	if _, ok := s.seen[normalizedURL]; ok {
		return false
	}
	s.seen[normalizedURL] = struct{}{}
	return true
}

func (s *websiteFetchState) appendChildIntels(out []Intel, childURLs []string, depth int) []Intel {
	if s == nil || depth <= 1 || len(childURLs) == 0 {
		return dedupWebsiteIntels(out)
	}
	for _, childURL := range childURLs {
		normalizedURL, err := normalizeWebsiteSeedURL(childURL)
		if err != nil {
			slog.Error("lacerate: child website normalize", "error", err, "url", childURL)
			continue
		}
		parsed, err := url.Parse(normalizedURL)
		if err != nil {
			slog.Error("lacerate: child website parse", "error", err, "url", childURL)
			continue
		}
		if isSkippableDirectMediaURL(parsed) {
			asset, err := directMediaFetchRoot(s.ctx, normalizedURL)
			if err != nil {
				slog.Error("lacerate: child media fetch", "error", err, "url", normalizedURL)
				continue
			}
			more, err := directMediaExtractAsset(s.ctx, s.db, 0, map[string]struct{}{}, &directMediaArchiveState{}, asset, Config.DirectMedia.MaxArchiveDepth)
			if err != nil {
				slog.Error("lacerate: child media extract", "error", err, "url", normalizedURL)
				continue
			}
			out = append(out, more...)
			continue
		}
		more, err := s.fetchers.FetchWebsite(normalizedURL, depth-1)
		if err != nil {
			slog.Error("lacerate: child website fetch", "error", err, "url", normalizedURL, "depth", depth-1)
			continue
		}
		out = append(out, more...)
	}
	return dedupWebsiteIntels(out)
}

func dedupWebsiteIntels(intels []Intel) []Intel {
	seen := map[string]struct{}{}
	out := make([]Intel, 0, len(intels))
	for i := range intels {
		if intels[i].DedupHash == nil || *intels[i].DedupHash == "" {
			continue
		}
		dedup := *intels[i].DedupHash
		if _, ok := seen[dedup]; ok {
			continue
		}
		seen[dedup] = struct{}{}
		out = append(out, intels[i])
	}
	return out
}

func (g *GeneralWebsiteFetcher) Priority(string) int {
	return 10
}

func (g *GeneralWebsiteFetcher) Match(rawURL string) bool {
	normalized, err := normalizeWebsiteSeedURL(rawURL)
	if err != nil {
		return false
	}
	parsed, err := url.Parse(normalized)
	if err != nil {
		return false
	}
	return !isSkippableDirectMediaURL(parsed)
}

func (g *GeneralWebsiteFetcher) FetchWebsite(rawURL string, depth int) ([]Intel, error) {
	if g.state == nil {
		return nil, fmt.Errorf("general website fetcher state is nil")
	}
	normalizedURL, err := normalizeWebsiteSeedURL(rawURL)
	if err != nil {
		return nil, err
	}
	if !g.state.markSeen(normalizedURL) {
		return nil, nil
	}
	page, err := fetchWebsitePage(g.state.ctx, normalizedURL)
	if err != nil {
		return nil, err
	}
	canonicalURL := websiteCanonicalURL(page.Doc, page.PageURL)
	if canonicalURL == "" {
		canonicalURL = normalizedURL
	}
	if canonicalURL != normalizedURL && !g.state.markSeen(canonicalURL) {
		return nil, nil
	}
	domain := page.PageURL.Scheme + "://" + page.PageURL.Host
	title := websitePageTitle(page.Doc)
	body := extractMarkdownFromFetchedHTML(g.state.ctx, page.HTML, page.PageURL, domain, true)
	intel, err := websiteIntelFromPage(title, body, canonicalURL, websitePublishedTime(page.Doc), g.state.now)
	if err != nil {
		return nil, err
	}
	out := []Intel{intel}
	childURLs := websiteLinksFromSelection(g.state.ctx, page.Doc.Selection, page.PageURL)
	return g.state.appendChildIntels(out, childURLs, depth), nil
}

func (m *MediumWebsiteScraper) Priority(string) int {
	return 100
}

func (m *MediumWebsiteScraper) Match(rawURL string) bool {
	normalized, err := normalizeWebsiteSeedURL(rawURL)
	if err != nil {
		return false
	}
	parsed, err := url.Parse(normalized)
	if err != nil {
		return false
	}
	host := strings.ToLower(parsed.Hostname())
	return host == "medium.com" || strings.HasSuffix(host, ".medium.com")
}

func (m *MediumWebsiteScraper) FetchWebsite(rawURL string, depth int) ([]Intel, error) {
	if m.state == nil {
		return nil, fmt.Errorf("medium website scraper state is nil")
	}
	normalizedURL, err := normalizeWebsiteSeedURL(rawURL)
	if err != nil {
		return nil, err
	}
	if !m.state.markSeen(normalizedURL) {
		return nil, nil
	}
	page, err := fetchWebsitePage(m.state.ctx, normalizedURL)
	if err != nil {
		return nil, err
	}
	canonicalURL := websiteCanonicalURL(page.Doc, page.PageURL)
	if canonicalURL == "" {
		canonicalURL = normalizedURL
	}
	if canonicalURL != normalizedURL && !m.state.markSeen(canonicalURL) {
		return nil, nil
	}
	article := page.Doc.Find("article").First()
	if article.Length() == 0 {
		return nil, fmt.Errorf("medium article element missing for %s", normalizedURL)
	}
	body := websiteSelectionMarkdown(article, page.PageURL)
	if markdownTooShort(body) {
		domain := page.PageURL.Scheme + "://" + page.PageURL.Host
		fallbackBody := extractMarkdownFromFetchedHTML(m.state.ctx, page.HTML, page.PageURL, domain, true)
		if !markdownTooShort(fallbackBody) || body == "" {
			body = fallbackBody
		}
	}
	title := websitePageTitle(page.Doc)
	intel, err := websiteIntelFromPage(title, body, canonicalURL, websitePublishedTime(page.Doc), m.state.now)
	if err != nil {
		return nil, err
	}
	out := []Intel{intel}
	childURLs := websiteLinksFromSelection(m.state.ctx, article, page.PageURL)
	return m.state.appendChildIntels(out, childURLs, depth), nil
}
