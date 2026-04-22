package p_seer_websites

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"strings"
	"time"

	"github.com/lariv-in/lago/lago"
	"golang.org/x/net/html"
	"gorm.io/gorm"
)

// Crawl caps to avoid runaway recursion or excessive Chromium load.
const (
	maxWebsiteSourceCrawlPages = 500
	maxWebsiteSourceDepth      = 16
)

// Fetch runs [FetchWebsiteSource] for this row (must be persisted with non-zero ID).
func (s *WebsiteSource) Fetch(ctx context.Context, db *gorm.DB) error {
	return FetchWebsiteSource(ctx, db, s)
}

// FetchWebsiteSource loads headless-rendered HTML for each URL in the crawl (breadth-limited by
// remaining depth), persists [Website] rows when markdown is extractable, and follows same-origin
// http(s) links until depth or caps are exhausted.
func FetchWebsiteSource(ctx context.Context, db *gorm.DB, src *WebsiteSource) error {
	if db == nil {
		return errors.New("p_seer_websites: FetchWebsiteSource: db is nil")
	}
	if src == nil || src.ID == 0 {
		return errors.New("p_seer_websites: FetchWebsiteSource: website source not loaded")
	}
	raw := strings.TrimSpace(src.URL.String())
	if raw == "" {
		return fmt.Errorf("p_seer_websites: website source %d has empty URL", src.ID)
	}

	if src.ID != 0 {
		_, loaded := websiteSourceCrawlBusy.LoadOrStore(src.ID, struct{}{})
		if loaded {
			return fmt.Errorf("p_seer_websites: crawl already running for website source %d", src.ID)
		}
		defer websiteSourceCrawlBusy.Delete(src.ID)
	}

	seed, err := fetchableWebsiteURL(ctx, raw)
	if err != nil {
		return err
	}

	maxDepth := src.Depth
	if maxDepth > maxWebsiteSourceDepth {
		maxDepth = maxWebsiteSourceDepth
	}

	type queued struct {
		raw       string
		levelLeft uint
	}

	queue := []queued{{raw: seed.String(), levelLeft: maxDepth}}
	seen := make(map[string]struct{}, 16)

	var crawlOrigin *url.URL
	var pagesProcessed int
	t0 := time.Now()

	for len(queue) > 0 && pagesProcessed < maxWebsiteSourceCrawlPages {
		job := queue[0]
		queue = queue[1:]

		canon, ferr := fetchableWebsiteURL(ctx, job.raw)
		if ferr != nil {
			slog.Warn("p_seer_websites: crawl skip URL", "error", ferr, "url", job.raw)
			continue
		}
		key := canon.String()
		if _, dup := seen[key]; dup {
			continue
		}
		seen[key] = struct{}{}

		htmlStr, finalU, err := fetchRenderedHTML(ctx, canon)
		if err != nil {
			slog.Warn("p_seer_websites: crawl fetch", "error", err, "url", key)
			continue
		}
		pagesProcessed++

		if crawlOrigin == nil && finalU != nil {
			o := *finalU
			o.Path, o.RawQuery, o.Fragment = "", "", ""
			crawlOrigin = cloneURL(&o)
		}

		if err := websiteCreateIfAbsentFromRenderedHTML(ctx, db, htmlStr, finalU); err != nil {
			slog.Warn("p_seer_websites: crawl persist website", "error", err, "url", key)
		}

		if job.levelLeft == 0 || crawlOrigin == nil || finalU == nil {
			continue
		}

		links := extractSameOriginLinks(ctx, htmlStr, finalU, crawlOrigin)
		nextLevel := job.levelLeft - 1
		for _, link := range links {
			if len(seen)+len(queue) >= maxWebsiteSourceCrawlPages {
				break
			}
			ls := link.String()
			if _, ok := seen[ls]; ok {
				continue
			}
			queue = append(queue, queued{raw: ls, levelLeft: nextLevel})
		}
	}

	if len(queue) > 0 && pagesProcessed >= maxWebsiteSourceCrawlPages {
		slog.Warn("p_seer_websites: crawl hit page cap",
			"website_source_id", src.ID,
			"cap", maxWebsiteSourceCrawlPages,
			"queued_remaining", len(queue),
		)
	}

	slog.Info("p_seer_websites: crawl summary",
		"website_source_id", src.ID,
		"seed_url", seed.String(),
		"depth_limit", maxDepth,
		"pages_fetched", pagesProcessed,
		"urls_seen", len(seen),
		"elapsed", time.Since(t0),
	)

	return nil
}

func sameOrigin(a, b *url.URL) bool {
	if a == nil || b == nil {
		return false
	}
	return strings.EqualFold(a.Scheme, b.Scheme) && strings.EqualFold(a.Host, b.Host)
}

func extractSameOriginLinks(ctx context.Context, htmlStr string, pageURL *url.URL, origin *url.URL) []*url.URL {
	if pageURL == nil || origin == nil {
		return nil
	}
	doc, err := html.Parse(strings.NewReader(htmlStr))
	if err != nil {
		return nil
	}
	var out []*url.URL
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "a" {
			for _, attr := range n.Attr {
				if attr.Key != "href" {
					continue
				}
				href := strings.TrimSpace(attr.Val)
				if href == "" || strings.HasPrefix(strings.ToLower(href), "javascript:") ||
					strings.HasPrefix(strings.ToLower(href), "mailto:") ||
					strings.HasPrefix(strings.ToLower(href), "tel:") {
					continue
				}
				ref, err := url.Parse(href)
				if err != nil {
					continue
				}
				abs := pageURL.ResolveReference(ref)
				if abs.Scheme != "http" && abs.Scheme != "https" {
					continue
				}
				canon, err := fetchableWebsiteURL(ctx, abs.String())
				if err != nil {
					continue
				}
				if !sameOrigin(canon, origin) {
					continue
				}
				out = append(out, cloneURL(canon))
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(doc)
	return dedupeURLPointers(out)
}

func dedupeURLPointers(in []*url.URL) []*url.URL {
	seen := make(map[string]struct{}, len(in))
	var out []*url.URL
	for _, u := range in {
		if u == nil {
			continue
		}
		s := u.String()
		if _, ok := seen[s]; ok {
			continue
		}
		seen[s] = struct{}{}
		out = append(out, u)
	}
	return out
}

func websiteCreateIfAbsentFromRenderedHTML(ctx context.Context, db *gorm.DB, fullHTML string, pageURL *url.URL) error {
	if db == nil {
		return errors.New("p_seer_websites: websiteCreateIfAbsentFromRenderedHTML: db is nil")
	}
	md := markdownFromRenderedHTML(fullHTML, pageURL)
	if strings.TrimSpace(md) == "" {
		return nil
	}
	canon := pageURL
	if canon == nil {
		return errors.New("p_seer_websites: websiteCreateIfAbsentFromRenderedHTML: page URL is nil")
	}
	keyU, err := fetchableWebsiteURL(ctx, canon.String())
	if err != nil {
		return err
	}
	key := keyU.String()

	var n int64
	if err := db.WithContext(ctx).Model(&Website{}).
		Where("url = ? AND deleted_at IS NULL", key).
		Count(&n).Error; err != nil {
		return fmt.Errorf("exists check: %w", err)
	}
	if n > 0 {
		return nil
	}

	var pp lago.PageURL
	pp.SetFromURL(keyU)
	w := Website{URL: pp, Markdown: md}
	if err := db.WithContext(ctx).Create(&w).Error; err != nil {
		return fmt.Errorf("create website: %w", err)
	}
	return nil
}
