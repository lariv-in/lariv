package p_lacerate

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

func parseRSSTime(s string) time.Time {
	s = strings.TrimSpace(s)
	for _, layout := range []string{
		time.RFC1123Z,
		time.RFC1123,
		time.RFC822Z,
		time.RFC822,
		time.RFC3339,
	} {
		if t, err := time.Parse(layout, s); err == nil {
			return t
		}
	}
	return time.Time{}
}

// twitterFetchedTweet is a normalized tweet for ingest (any fetch mode).
type twitterFetchedTweet struct {
	ID        string
	Text      string
	CreatedAt time.Time
	Permalink string
	ImageURL  string
}

const twitterHTTPTimeout = 60 * time.Second

var statusPathRE = regexp.MustCompile(`/status/(\d+)`)

func fetchTweetsForHandle(ctx context.Context, handle string) ([]twitterFetchedTweet, error) {
	if Config == nil {
		err := fmt.Errorf("lacerate: p_lacerate config not initialized")
		slog.Error("lacerate: twitter fetch tweets", "error", err, "handle", handle)
		return nil, err
	}
	switch Config.Twitter.FetchMode {
	case TwitterFetchBearer:
		return fetchTwitterBearer(ctx, handle)
	case TwitterFetchNitter:
		return fetchTwitterNitterRSS(ctx, handle)
	case TwitterFetchScraping:
		return fetchTwitterScrapeNitterHTML(ctx, handle)
	default:
		err := fmt.Errorf("lacerate: twitter.fetchMode not configured (add [plugins.p_lacerate] with twitter.fetchMode to totschool.toml)")
		slog.Error("lacerate: twitter fetch tweets", "error", err, "handle", handle)
		return nil, err
	}
}

func twitterHTTPClient() *http.Client {
	return &http.Client{Timeout: twitterHTTPTimeout}
}

// --- Twitter API v2 (Bearer) ---

type twitterAPIUserResp struct {
	Data *struct {
		ID string `json:"id"`
	} `json:"data"`
	Errors []json.RawMessage `json:"errors"`
}

type twitterAPITweetsResp struct {
	Data     []twitterAPITweet    `json:"data"`
	Includes *twitterAPIIncludes  `json:"includes"`
	Meta     twitterAPITweetsMeta `json:"meta"`
	Errors   []json.RawMessage    `json:"errors"`
}

type twitterAPITweet struct {
	ID          string `json:"id"`
	Text        string `json:"text"`
	CreatedAt   string `json:"created_at"`
	Attachments *struct {
		MediaKeys []string `json:"media_keys"`
	} `json:"attachments"`
}

type twitterAPIIncludes struct {
	Media []struct {
		Type            string `json:"type"`
		URL             string `json:"url"`
		PreviewImageURL string `json:"preview_image_url"`
	} `json:"media"`
}

type twitterAPITweetsMeta struct {
	NextToken   string `json:"next_token"`
	ResultCount int    `json:"result_count"`
}

func fetchTwitterBearer(ctx context.Context, handle string) ([]twitterFetchedTweet, error) {
	userID, err := twitterAPIUserIDByUsername(ctx, handle)
	if err != nil {
		return nil, err
	}
	token := strings.TrimSpace(Config.Twitter.BearerToken)

	type mediaObj struct {
		MediaKey        string `json:"media_key"`
		Type            string `json:"type"`
		URL             string `json:"url"`
		PreviewImageURL string `json:"preview_image_url"`
	}

	const maxPages = 25
	var out []twitterFetchedTweet
	next := ""

	for range maxPages {
		u, err := url.Parse("https://api.twitter.com/2/users/" + url.PathEscape(userID) + "/tweets")
		if err != nil {
			slog.Error("lacerate: twitter api tweets url", "error", err, "handle", handle)
			return nil, err
		}
		q := u.Query()
		q.Set("max_results", "100")
		q.Set("tweet.fields", "created_at,attachments")
		q.Set("expansions", "attachments.media_keys")
		q.Set("media.fields", "type,url,preview_image_url")
		if next != "" {
			q.Set("pagination_token", next)
		}
		u.RawQuery = q.Encode()

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
		if err != nil {
			slog.Error("lacerate: twitter api tweets request", "error", err, "handle", handle)
			return nil, err
		}
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err := twitterHTTPClient().Do(req)
		if err != nil {
			slog.Error("lacerate: twitter api tweets", "error", err, "handle", handle)
			return nil, err
		}
		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			slog.Error("lacerate: twitter api tweets read body", "error", err, "handle", handle)
			return nil, err
		}
		if resp.StatusCode != http.StatusOK {
			err := fmt.Errorf("twitter api tweets: %s: %s", resp.Status, strings.TrimSpace(string(body)))
			slog.Error("lacerate: twitter api tweets", "error", err, "handle", handle)
			return nil, err
		}

		var tr twitterAPITweetsResp
		if err := json.Unmarshal(body, &tr); err != nil {
			slog.Error("lacerate: twitter api tweets decode", "error", err, "handle", handle)
			return nil, err
		}

		var rawWrap struct {
			Includes struct {
				Media []mediaObj `json:"media"`
			} `json:"includes"`
		}
		if err := json.Unmarshal(body, &rawWrap); err != nil {
			slog.Error("lacerate: twitter api tweets decode includes", "error", err, "handle", handle)
			return nil, err
		}
		mediaByKey := map[string]mediaObj{}
		for _, m := range rawWrap.Includes.Media {
			if m.MediaKey != "" {
				mediaByKey[m.MediaKey] = m
			}
		}

		for _, tw := range tr.Data {
			created, _ := time.Parse(time.RFC3339, tw.CreatedAt)
			permalink := fmt.Sprintf("https://x.com/%s/status/%s", handle, tw.ID)
			ft := twitterFetchedTweet{
				ID:        tw.ID,
				Text:      tw.Text,
				CreatedAt: created,
				Permalink: permalink,
			}
			if tw.Attachments != nil {
				for _, mk := range tw.Attachments.MediaKeys {
					if mo, ok := mediaByKey[mk]; ok {
						switch mo.Type {
						case "photo":
							if mo.URL != "" {
								ft.ImageURL = mo.URL
							} else if mo.PreviewImageURL != "" {
								ft.ImageURL = mo.PreviewImageURL
							}
						default:
							if mo.PreviewImageURL != "" {
								ft.ImageURL = mo.PreviewImageURL
							}
						}
						if ft.ImageURL != "" {
							break
						}
					}
				}
			}
			out = append(out, ft)
		}

		if tr.Meta.NextToken == "" {
			break
		}
		next = tr.Meta.NextToken
	}

	return out, nil
}

func twitterAPIUserIDByUsername(ctx context.Context, handle string) (string, error) {
	u := "https://api.twitter.com/2/users/by/username/" + url.PathEscape(handle)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		slog.Error("lacerate: twitter api user lookup request", "error", err, "handle", handle)
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+strings.TrimSpace(Config.Twitter.BearerToken))

	resp, err := twitterHTTPClient().Do(req)
	if err != nil {
		slog.Error("lacerate: twitter api user lookup", "error", err, "handle", handle)
		return "", err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		slog.Error("lacerate: twitter api user lookup read body", "error", err, "handle", handle)
		return "", err
	}
	if resp.StatusCode != http.StatusOK {
		err := fmt.Errorf("twitter api user lookup: %s: %s", resp.Status, strings.TrimSpace(string(body)))
		slog.Error("lacerate: twitter api user lookup", "error", err, "handle", handle)
		return "", err
	}
	var ur twitterAPIUserResp
	if err := json.Unmarshal(body, &ur); err != nil {
		slog.Error("lacerate: twitter api user lookup decode", "error", err, "handle", handle)
		return "", err
	}
	if ur.Data == nil || ur.Data.ID == "" {
		err := fmt.Errorf("twitter api: user %q not found", handle)
		slog.Error("lacerate: twitter api user lookup", "error", err, "handle", handle)
		return "", err
	}
	return ur.Data.ID, nil
}

// --- Nitter RSS ---

type nitterRSS struct {
	XMLName xml.Name         `xml:"rss"`
	Channel nitterRSSChannel `xml:"channel"`
}

type nitterRSSChannel struct {
	Items []nitterRSSItem `xml:"item"`
}

type nitterRSSItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	GUID        string `xml:"guid"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
	Enclosure   *struct {
		URL string `xml:"url,attr"`
	} `xml:"enclosure"`
}

func fetchTwitterNitterRSS(ctx context.Context, handle string) ([]twitterFetchedTweet, error) {
	base := Config.Twitter.NitterBaseURL
	rssURL := base + "/" + url.PathEscape(handle) + "/rss"

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rssURL, nil)
	if err != nil {
		slog.Error("lacerate: nitter rss request", "error", err, "handle", handle)
		return nil, err
	}
	req.Header.Set("User-Agent", Config.IntelPreview.UserAgent)
	req.Header.Set("Accept", "application/rss+xml, application/xml, text/xml;q=0.9, */*;q=0.8")

	resp, err := twitterHTTPClient().Do(req)
	if err != nil {
		slog.Error("lacerate: nitter rss fetch", "error", err, "handle", handle)
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		slog.Error("lacerate: nitter rss read body", "error", err, "handle", handle)
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		err := fmt.Errorf("nitter rss: %s", resp.Status)
		slog.Error("lacerate: nitter rss", "error", err, "handle", handle)
		return nil, err
	}

	var feed nitterRSS
	if err := xml.Unmarshal(body, &feed); err != nil {
		err := fmt.Errorf("nitter rss parse: %w", err)
		slog.Error("lacerate: nitter rss parse", "error", err, "handle", handle)
		return nil, err
	}

	var out []twitterFetchedTweet
	for _, it := range feed.Channel.Items {
		stable := strings.TrimSpace(it.GUID)
		if stable == "" {
			stable = strings.TrimSpace(it.Link)
		}
		id := twitterStableIDFromLinkOrGUID(stable)
		if id == "" {
			continue
		}
		created := parseRSSTime(it.PubDate)
		text := strings.TrimSpace(it.Title)
		if text == "" {
			text = strings.TrimSpace(stripHTMLSnippet(it.Description))
		}
		link := strings.TrimSpace(it.Link)
		ft := twitterFetchedTweet{
			ID:        id,
			Text:      text,
			CreatedAt: created,
			Permalink: link,
		}
		if it.Enclosure != nil && strings.TrimSpace(it.Enclosure.URL) != "" {
			ft.ImageURL = strings.TrimSpace(it.Enclosure.URL)
		}
		out = append(out, ft)
	}
	return out, nil
}

func twitterStableIDFromLinkOrGUID(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	if m := statusPathRE.FindStringSubmatch(s); len(m) == 2 {
		return m[1]
	}
	// fallback: hash-length safe id from full string
	if len(s) > 80 {
		return s[len(s)-40:]
	}
	return s
}

func stripHTMLSnippet(s string) string {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(s))
	if err != nil {
		slog.Warn("lacerate: strip html snippet", "error", err)
		return s
	}
	return strings.TrimSpace(doc.Text())
}

// --- Nitter HTML + Rod (scraping) ---

// fetchTwitterScrapeNitterHTML loads a Nitter profile timeline in headless Chrome and parses tweet cards.
// Selectors target common Nitter layouts; instances differ — adjust tweetTimelineItemSelectors if needed.
var tweetTimelineItemSelectors = []string{".timeline-item", "div.timeline-item", ".timeline .timeline-item"}

func fetchTwitterScrapeNitterHTML(ctx context.Context, handle string) ([]twitterFetchedTweet, error) {
	base := Config.Twitter.NitterBaseURL
	pageURL := base + "/" + url.PathEscape(handle)

	html, err := fetchHTMLViaRod(ctx, pageURL)
	if err != nil {
		slog.Error("lacerate: nitter scrape fetch html", "error", err, "handle", handle)
		return nil, err
	}
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		slog.Error("lacerate: nitter scrape parse html", "error", err, "handle", handle)
		return nil, err
	}

	var items *goquery.Selection
	for _, sel := range tweetTimelineItemSelectors {
		found := doc.Find(sel)
		if found.Length() > 0 {
			items = found
			break
		}
	}
	if items == nil || items.Length() == 0 {
		err := fmt.Errorf("nitter scrape: no timeline items (layout may have changed; try fetchMode %q)", TwitterFetchNitter)
		slog.Error("lacerate: nitter scrape", "error", err, "handle", handle)
		return nil, err
	}

	baseU, err := url.Parse(base)
	if err != nil {
		slog.Error("lacerate: nitter scrape base url", "error", err, "handle", handle)
		return nil, err
	}

	var out []twitterFetchedTweet
	items.Each(func(_ int, s *goquery.Selection) {
		linkEl := s.Find("a.tweet-link").First()
		if linkEl.Length() == 0 {
			linkEl = s.Find("a[href*='/status/']").First()
		}
		href, _ := linkEl.Attr("href")
		href = strings.TrimSpace(href)
		if href == "" {
			return
		}
		permalink := href
		if strings.HasPrefix(permalink, "/") {
			permalink = baseU.ResolveReference(&url.URL{Path: permalink}).String()
		}
		id := twitterStableIDFromLinkOrGUID(href)
		if id == "" {
			return
		}
		text := strings.TrimSpace(s.Find(".tweet-content").First().Text())
		if text == "" {
			text = strings.TrimSpace(s.Find(".tweet-body").First().Text())
		}
		if text == "" {
			return
		}
		out = append(out, twitterFetchedTweet{
			ID:        id,
			Text:      text,
			CreatedAt: time.Time{},
			Permalink: permalink,
		})
	})

	if len(out) == 0 {
		err := fmt.Errorf("nitter scrape: found timeline markup but no parseable tweets")
		slog.Error("lacerate: nitter scrape", "error", err, "handle", handle)
		return nil, err
	}
	return out, nil
}
