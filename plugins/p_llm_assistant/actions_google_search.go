package p_llm_assistant

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const googleAssistantCSEEndpoint = "https://www.googleapis.com/customsearch/v1"

const googleAssistantCSEPageSize = 10

type assistantGoogleCSEItem struct {
	Title   string `json:"title"`
	Link    string `json:"link"`
	Snippet string `json:"snippet"`
}

type assistantGoogleCSEResponse struct {
	Items []assistantGoogleCSEItem `json:"items"`
}

type assistantGoogleHit struct {
	Title   string `json:"title"`
	Link    string `json:"link"`
	Snippet string `json:"snippet"`
}

// runGoogleSearchTool runs Google Programmable Search (Custom Search JSON API).
func runGoogleSearchTool(ctx context.Context, query string, limit int) ([]assistantGoogleHit, error) {
	cfg := LlmAssistantPlugin
	if cfg == nil {
		return nil, fmt.Errorf("google_search: config unavailable")
	}
	key := cfg.CseAPIKey
	cx := cfg.CseCX
	if key == "" || cx == "" {
		return nil, fmt.Errorf(`google_search: configure [Plugins.p_llm_assistant] cseApiKey and cseCx (same Google Custom Search credentials as programmable search)`)
	}
	q := strings.TrimSpace(query)
	if q == "" {
		return nil, fmt.Errorf("google_search: empty query")
	}
	if limit <= 0 {
		limit = 8
	}
	cap := AssistantAppConfig.GoogleSearchResultLimitCap
	if cap <= 0 {
		cap = defaultGoogleSearchResultLimitCap
	}
	if limit > cap {
		limit = cap
	}

	client := &http.Client{Timeout: 30 * time.Second}
	hits := make([]assistantGoogleHit, 0, limit)
	maxPages := (limit + googleAssistantCSEPageSize - 1) / googleAssistantCSEPageSize
	if maxPages < 1 {
		maxPages = 1
	}
	if maxPages > assistantGoogleSearchMaxPages {
		maxPages = assistantGoogleSearchMaxPages
	}
	for page := 0; len(hits) < limit && page < maxPages; page++ {
		need := limit - len(hits)
		n := googleAssistantCSEPageSize
		if need < n {
			n = need
		}
		start := 1 + page*googleAssistantCSEPageSize
		u, err := url.Parse(googleAssistantCSEEndpoint)
		if err != nil {
			return nil, err
		}
		qv := url.Values{}
		qv.Set("key", key)
		qv.Set("cx", cx)
		qv.Set("q", q)
		qv.Set("num", strconv.Itoa(n))
		qv.Set("start", strconv.Itoa(start))
		u.RawQuery = qv.Encode()

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
		if err != nil {
			return nil, err
		}
		res, err := client.Do(req)
		if err != nil {
			return nil, err
		}
		body, err := io.ReadAll(io.LimitReader(res.Body, 8<<20))
		_ = res.Body.Close()
		if err != nil {
			return nil, err
		}
		if res.StatusCode != http.StatusOK {
			slog.Warn("p_llm_assistant: CSE HTTP error", "status", res.StatusCode, "body", truncateGoogleSearchLog(string(body)))
			return nil, fmt.Errorf("google_search: HTTP status %d", res.StatusCode)
		}
		var parsed assistantGoogleCSEResponse
		if err := json.Unmarshal(body, &parsed); err != nil {
			return nil, fmt.Errorf("google_search: response json: %v", err)
		}
		if len(parsed.Items) == 0 {
			break
		}
		for _, it := range parsed.Items {
			link := strings.TrimSpace(it.Link)
			if link == "" {
				continue
			}
			hits = append(hits, assistantGoogleHit{
				Title:   strings.TrimSpace(it.Title),
				Link:    link,
				Snippet: strings.TrimSpace(it.Snippet),
			})
			if len(hits) >= limit {
				break
			}
		}
		if len(parsed.Items) < n {
			break
		}
	}
	return hits, nil
}

func truncateGoogleSearchLog(s string) string {
	if len(s) <= 400 {
		return s
	}
	return s[:400] + "…"
}
