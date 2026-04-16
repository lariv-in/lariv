package p_lacerate

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"

	"gorm.io/gorm"
)

const (
	googleSearchEndpoint             = "https://www.googleapis.com/customsearch/v1"
	googleSearchHTTPTimeout          = 20 * time.Second
	googleSearchSourceDefaultResults = 10
)

type googleSearchResponse struct {
	Items []googleSearchResult `json:"items"`
	Error *struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

type googleSearchResult struct {
	Title       string `json:"title"`
	Link        string `json:"link"`
	Snippet     string `json:"snippet"`
	DisplayLink string `json:"displayLink"`
	Pagemap     struct {
		CSEImage []struct {
			Src string `json:"src"`
		} `json:"cse_image"`
		CSEThumbnail []struct {
			Src string `json:"src"`
		} `json:"cse_thumbnail"`
	} `json:"pagemap"`
}

func googleSearchQuery(ctx context.Context, apiKey, cx, query string, num int) ([]googleSearchResult, error) {
	apiKey = strings.TrimSpace(apiKey)
	cx = strings.TrimSpace(cx)
	query = strings.TrimSpace(query)
	if apiKey == "" {
		return nil, fmt.Errorf("google search api key is not configured (p_lacerate.googleSearch.apiKey)")
	}
	if cx == "" {
		return nil, fmt.Errorf("google search cx is not configured (p_lacerate.googleSearch.cx)")
	}
	if query == "" {
		return nil, fmt.Errorf("google search query is empty")
	}
	if num <= 0 {
		num = googleSearchSourceDefaultResults
	}

	u, err := url.Parse(googleSearchEndpoint)
	if err != nil {
		return nil, err
	}
	q := u.Query()
	q.Set("key", apiKey)
	q.Set("cx", cx)
	q.Set("q", query)
	q.Set("num", fmt.Sprintf("%d", num))
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}
	client := &http.Client{Timeout: googleSearchHTTPTimeout}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("google search request: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
	if err != nil {
		return nil, err
	}

	var gr googleSearchResponse
	if err := json.Unmarshal(body, &gr); err != nil {
		slog.Error("lacerate: google search json parse", "error", err)
		return nil, fmt.Errorf("google search response: invalid json")
	}
	if resp.StatusCode != http.StatusOK {
		msg := strings.TrimSpace(resp.Status)
		if gr.Error != nil && strings.TrimSpace(gr.Error.Message) != "" {
			msg = strings.TrimSpace(gr.Error.Message)
		}
		return nil, fmt.Errorf("google search failed: %s", msg)
	}
	return gr.Items, nil
}

func googleSearchResultDedupHash(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	sum := sha256.Sum256([]byte("google_search_result:" + raw))
	return hex.EncodeToString(sum[:])
}

func googleSearchBuildContent(item googleSearchResult, link string) string {
	var b strings.Builder
	title := strings.TrimSpace(item.Title)
	snippet := strings.TrimSpace(item.Snippet)
	displayLink := strings.TrimSpace(item.DisplayLink)
	if title != "" {
		fmt.Fprintf(&b, "## %s\n\n", title)
	}
	if snippet != "" {
		b.WriteString(snippet)
		b.WriteString("\n\n")
	}
	b.WriteString("---\n\n")
	if link != "" {
		fmt.Fprintf(&b, "- **Link:** %s\n", link)
	}
	if displayLink != "" {
		fmt.Fprintf(&b, "- **Display Link:** %s\n", displayLink)
	}
	return strings.TrimSpace(b.String())
}

func googleSearchPreviewURL(item googleSearchResult) string {
	for _, img := range item.Pagemap.CSEThumbnail {
		src := strings.TrimSpace(img.Src)
		if src != "" {
			return src
		}
	}
	for _, img := range item.Pagemap.CSEImage {
		src := strings.TrimSpace(img.Src)
		if src != "" {
			return src
		}
	}
	return ""
}

func runWebsearchQueryFetch(ctx context.Context, db *gorm.DB, sourceID *uint, query string, existingDedup map[string]struct{}) ([]Intel, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return nil, fmt.Errorf("websearch query is empty")
	}
	if existingDedup == nil {
		existingDedup = map[string]struct{}{}
	}
	items, err := googleSearchQuery(ctx, Config.GoogleSearch.APIKey, Config.GoogleSearch.CX, query, googleSearchSourceDefaultResults)
	if err != nil {
		return nil, err
	}
	if len(items) == 0 {
		return nil, nil
	}

	fetchers := NewWebsiteFetchers(ctx, db)
	out := make([]Intel, 0, len(items))

	for _, item := range items {
		rawLink := strings.TrimSpace(item.Link)
		crawlLink := ""
		if normalized, err := normalizeWebsiteSeedURL(rawLink); err == nil {
			crawlLink = normalized
		} else if rawLink != "" {
			slog.Warn("lacerate: websearch result link not crawlable", "error", err, "link", rawLink, "query", query)
		}

		if crawlLink == "" {
			continue
		}
		crawled, err := fetchers.FetchWebsite(crawlLink, 1)
		if err != nil {
			slog.Error("lacerate: websearch crawl result", "error", err, "link", crawlLink, "query", query)
			continue
		}
		for i := range crawled {
			dh := crawled[i].DedupHash
			if dh == nil || *dh == "" {
				slog.Warn("lacerate: websearch crawled intel missing dedupe", "link", crawlLink, "query", query)
				continue
			}
			if _, dup := existingDedup[*dh]; dup {
				continue
			}
			crawled[i].SourceID = sourceID
			out = append(out, crawled[i])
			existingDedup[*dh] = struct{}{}
		}
	}

	return out, nil
}
