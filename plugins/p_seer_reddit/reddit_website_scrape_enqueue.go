package p_seer_reddit

import (
	"net/url"
	"regexp"
	"strings"

	"github.com/lariv-in/lago/plugins/p_seer_websites"
)

var redditPostHTTPURLRe = regexp.MustCompile(`https?://[^\s<>()\[\]"']+`)

func websiteScrapeHostSkipped(host string) bool {
	h := strings.ToLower(strings.TrimSpace(host))
	if h == "" {
		return true
	}
	if h == "reddit.com" || strings.HasSuffix(h, ".reddit.com") {
		return true
	}
	return false
}

func tryEnqueueWebsiteScrapeURL(raw string) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return
	}
	parsed, err := url.Parse(raw)
	if err != nil || parsed.Host == "" {
		return
	}
	switch strings.ToLower(parsed.Scheme) {
	case "http", "https":
	default:
		return
	}
	if websiteScrapeHostSkipped(parsed.Hostname()) {
		return
	}
	toSend, err := url.Parse(parsed.String())
	if err != nil || toSend.Host == "" {
		return
	}
	p_seer_websites.WebsiteScrapeURLQueue <- toSend
}

func dedupeStrings(in []string) []string {
	seen := make(map[string]struct{}, len(in))
	var out []string
	for _, s := range in {
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}
		if _, ok := seen[s]; ok {
			continue
		}
		seen[s] = struct{}{}
		out = append(out, s)
	}
	return out
}

// enqueueURLsFromRedditPost pushes external http(s) URLs from link-post [RedditPostData.URL]
// and from http(s) matches in title/selftext to [p_seer_websites.WebsiteScrapeURLQueue].
// Call only when [RedditSource.LoadWebsites] is true ([persistPostIfNew] gate).
func enqueueURLsFromRedditPost(post RedditPostData) {
	var raw []string
	if u := strings.TrimSpace(post.URL); u != "" {
		raw = append(raw, u)
	}
	scan := post.Title + "\n" + post.Selftext
	for _, m := range redditPostHTTPURLRe.FindAllString(scan, -1) {
		raw = append(raw, strings.TrimRight(m, ".,;:!?)"))
	}
	for _, s := range dedupeStrings(raw) {
		tryEnqueueWebsiteScrapeURL(s)
	}
}
