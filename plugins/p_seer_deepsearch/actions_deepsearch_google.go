package p_seer_deepsearch

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

const (
	googleCSEEndpoint     = "https://www.googleapis.com/customsearch/v1"
	googleCSEPageSize     = 10
	maxGoogleCSEPagesPerQ = 2
)

type googleCSEItem struct {
	Link string `json:"link"`
}

type googleCSEResponse struct {
	Items []googleCSEItem `json:"items"`
}

// googleCustomSearchURLs returns organic result links for one CSE query (paginated up to [maxGoogleCSEPagesPerQ] pages).
func googleCustomSearchURLs(ctx context.Context, searchQuery string) ([]string, error) {
	key := strings.TrimSpace(DeepSearchAppConfig.APIKey)
	cx := strings.TrimSpace(DeepSearchAppConfig.CX)
	if key == "" || cx == "" {
		return nil, fmt.Errorf("p_seer_deepsearch: Google CSE apiKey or cx is empty")
	}
	q := strings.TrimSpace(searchQuery)
	if q == "" {
		return nil, nil
	}

	client := &http.Client{Timeout: 30 * time.Second}
	var out []string
	for page := 0; page < maxGoogleCSEPagesPerQ; page++ {
		start := 1 + page*googleCSEPageSize
		u, err := url.Parse(googleCSEEndpoint)
		if err != nil {
			return nil, err
		}
		qv := url.Values{}
		qv.Set("key", key)
		qv.Set("cx", cx)
		qv.Set("q", q)
		qv.Set("num", strconv.Itoa(googleCSEPageSize))
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
			slog.Warn("p_seer_deepsearch: CSE HTTP error", "status", res.StatusCode, "body", truncateForLog(string(body)))
			return nil, fmt.Errorf("p_seer_deepsearch: CSE status %d", res.StatusCode)
		}
		var parsed googleCSEResponse
		if err := json.Unmarshal(body, &parsed); err != nil {
			return nil, fmt.Errorf("p_seer_deepsearch: CSE json: %w", err)
		}
		if len(parsed.Items) == 0 {
			break
		}
		for _, it := range parsed.Items {
			if s := strings.TrimSpace(it.Link); s != "" {
				out = append(out, s)
			}
		}
		if len(parsed.Items) < googleCSEPageSize {
			break
		}
	}
	return out, nil
}

func truncateForLog(s string) string {
	if len(s) <= 400 {
		return s
	}
	return s[:400] + "…"
}
