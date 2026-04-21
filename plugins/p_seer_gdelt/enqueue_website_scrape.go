package p_seer_gdelt

import (
	"log/slog"
	"net/url"
	"strings"

	"github.com/lariv-in/lago/plugins/p_seer_websites"
)

// EnqueueEventSourceURLForWebsiteScrape sends [Event.SourceURL] to
// [p_seer_websites.WebsiteScrapeURLQueue] so the Seer Websites worker can scrape
// the article page when it is not already stored. Non-blocking: if the queue is
// full, the URL is dropped and a warning is logged.
func EnqueueEventSourceURLForWebsiteScrape(raw string) {
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
	toSend, err := url.Parse(parsed.String())
	if err != nil || toSend.Host == "" {
		return
	}
	select {
	case p_seer_websites.WebsiteScrapeURLQueue <- toSend:
	default:
		slog.Warn("p_seer_gdelt: website scrape URL queue full", "url", raw)
	}
}
