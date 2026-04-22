package p_seer_intel

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/lariv-in/lago/plugins/p_google_genai"
	"gorm.io/gorm"
)

const intelEventExtractSystemPrompt = `You extract a geographic location and an event time from an intelligence summary.
Respond ONLY with a JSON object (no markdown, no commentary) with exactly these keys:
"address": string — a concise postal-style or place name string suitable for a geocoder (empty if none).
"datetime": string — RFC3339 timestamp in UTC (e.g. "2006-01-02T15:04:05Z").
If the summary implies a date but not a time, use 00:00:00Z for that date.
If no date can be inferred, use the current UTC time in RFC3339.`

const (
	geocodeHTTPTimeout      = 15 * time.Second
	geocodeRetryMaxAttempts = 6
)

type intelEventLLMOut struct {
	Address  string `json:"address"`
	Datetime string `json:"datetime"`
}

// extractIntelEventFromSummary asks Gemini ([p_google_genai]) for a JSON object containing address + event time.
func extractIntelEventFromSummary(ctx context.Context, summary string) (address string, eventTime time.Time, err error) {
	summary = strings.TrimSpace(summary)
	if summary == "" {
		return "", time.Time{}, fmt.Errorf("p_seer_intel: extract intel event: empty summary")
	}
	userPrompt := fmt.Sprintf("Current UTC time: %s\n\nSummary:\n%s", time.Now().UTC().Format(time.RFC3339), summary)
	var out intelEventLLMOut
	raw, err := p_google_genai.GenerateJSON(ctx, p_google_genai.GenerateRequest{
		SystemPrompt:    intelEventExtractSystemPrompt,
		UserPrompt:      userPrompt,
		MaxOutputTokens: 256,
		Thinking:        &p_google_genai.ThinkingConfig{Mode: p_google_genai.ThinkingModeDisabled},
	}, &out)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("p_seer_intel: event extract generate: %w", err)
	}
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", time.Time{}, fmt.Errorf("p_seer_intel: event extract returned empty")
	}
	eventTime, err = time.Parse(time.RFC3339, strings.TrimSpace(out.Datetime))
	if err != nil {
		return "", time.Time{}, fmt.Errorf("p_seer_intel: event extract datetime: %w", err)
	}
	return strings.TrimSpace(out.Address), eventTime.UTC(), nil
}

type googleGeocodeResponse struct {
	Status  string `json:"status"`
	Results []struct {
		Geometry struct {
			Location struct {
				Lat float64 `json:"lat"`
				Lng float64 `json:"lng"`
			} `json:"location"`
		} `json:"geometry"`
	} `json:"results"`
}

func geocodeAPIStatusRetryable(status string) bool {
	switch status {
	case "OVER_QUERY_LIMIT", "UNKNOWN_ERROR":
		return true
	default:
		return false
	}
}

func geocodeHTTPStatusRetryable(code int) bool {
	switch code {
	case http.StatusTooManyRequests, http.StatusInternalServerError,
		http.StatusBadGateway, http.StatusServiceUnavailable, http.StatusGatewayTimeout:
		return true
	default:
		return false
	}
}

func geocodeBackoffDelay(attempt int) time.Duration {
	if attempt < 0 {
		attempt = 0
	}
	shift := attempt
	if shift > 12 {
		shift = 12
	}
	d := 400 * time.Millisecond * time.Duration(int64(1)<<shift)
	if d > 60*time.Second || d <= 0 {
		d = 60 * time.Second
	}
	jcap := d / 4
	if jcap > 2*time.Second {
		jcap = 2 * time.Second
	}
	if jcap <= 0 {
		return d
	}
	return d + time.Duration(rand.Int63n(int64(jcap)+1))
}

// geocodeGoogleMapsOnce performs one Geocoding HTTP GET. retry is true when the caller should backoff and try again.
func geocodeGoogleMapsOnce(ctx context.Context, requestURL string) (lat, lng float64, retry bool, err error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL, nil)
	if err != nil {
		return 0, 0, false, err
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			return 0, 0, false, err
		}
		return 0, 0, true, fmt.Errorf("p_seer_intel: geocode http: %w", err)
	}
	defer res.Body.Close()
	body, err := io.ReadAll(io.LimitReader(res.Body, 4<<20))
	if err != nil {
		return 0, 0, true, fmt.Errorf("p_seer_intel: geocode read body: %w", err)
	}
	if geocodeHTTPStatusRetryable(res.StatusCode) {
		return 0, 0, true, fmt.Errorf("p_seer_intel: geocode http status %d", res.StatusCode)
	}
	if res.StatusCode != http.StatusOK {
		return 0, 0, false, fmt.Errorf("p_seer_intel: geocode http status %d", res.StatusCode)
	}
	var parsed googleGeocodeResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		return 0, 0, false, fmt.Errorf("p_seer_intel: geocode json: %w", err)
	}
	if parsed.Status == "OK" && len(parsed.Results) > 0 {
		loc := parsed.Results[0].Geometry.Location
		return loc.Lat, loc.Lng, false, nil
	}
	if geocodeAPIStatusRetryable(parsed.Status) {
		return 0, 0, true, fmt.Errorf("p_seer_intel: geocode status %q", parsed.Status)
	}
	return 0, 0, false, fmt.Errorf("p_seer_intel: geocode status %q", parsed.Status)
}

// geocodeGoogleMaps returns latitude and longitude for address using the Geocoding API.
// Retries transient HTTP failures and Google statuses OVER_QUERY_LIMIT / UNKNOWN_ERROR with
// truncated exponential backoff + jitter.
func geocodeGoogleMaps(ctx context.Context, apiKey, address string) (lat, lng float64, err error) {
	address = strings.TrimSpace(address)
	apiKey = strings.TrimSpace(apiKey)
	if address == "" || apiKey == "" {
		return 0, 0, fmt.Errorf("p_seer_intel: geocode: empty address or key")
	}

	u, err := url.Parse("https://maps.googleapis.com/maps/api/geocode/json")
	if err != nil {
		return 0, 0, err
	}
	q := u.Query()
	q.Set("address", address)
	q.Set("key", apiKey)
	u.RawQuery = q.Encode()
	requestURL := u.String()

	for attempt := 0; attempt < geocodeRetryMaxAttempts; attempt++ {
		attemptCtx, cancel := context.WithTimeout(ctx, geocodeHTTPTimeout)
		lat, lng, retry, err := geocodeGoogleMapsOnce(attemptCtx, requestURL)
		cancel()
		if err == nil {
			return lat, lng, nil
		}
		if !retry || attempt >= geocodeRetryMaxAttempts-1 {
			return 0, 0, err
		}
		wait := geocodeBackoffDelay(attempt)
		slog.Warn("p_seer_intel: geocode retry", "attempt", attempt+1, "max", geocodeRetryMaxAttempts, "wait", wait, "err", err)
		timer := time.NewTimer(wait)
		select {
		case <-ctx.Done():
			if !timer.Stop() {
				<-timer.C
			}
			return 0, 0, ctx.Err()
		case <-timer.C:
		}
	}
	return 0, 0, fmt.Errorf("p_seer_intel: geocode: unreachable")
}

// CreateIntelAndEvent persists intel, then best-effort creates [IntelEvent] from [Intel.Summary] (LLM + geocode on insert).
// Returns an error only if saving [Intel] fails.
func CreateIntelAndEvent(ctx context.Context, db *gorm.DB, intel *Intel) error {
	if db == nil {
		return fmt.Errorf("p_seer_intel: CreateIntelAndEvent: db is nil")
	}
	if intel == nil {
		return fmt.Errorf("p_seer_intel: CreateIntelAndEvent: intel is nil")
	}
	if err := db.WithContext(ctx).Create(intel).Error; err != nil {
		return err
	}
	addr, eventTime, err := extractIntelEventFromSummary(ctx, intel.Summary)
	if err != nil {
		slog.Warn("p_seer_intel: intel event extract skipped", "intel_id", intel.ID, "error", err)
		return nil
	}
	ev := IntelEvent{
		IntelID:  intel.ID,
		Address:  addr,
		Datetime: eventTime,
	}
	if err := db.WithContext(ctx).Create(&ev).Error; err != nil {
		slog.Warn("p_seer_intel: intel event persist failed", "intel_id", intel.ID, "error", err)
	}
	return nil
}
