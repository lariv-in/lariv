package p_lacerate

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const googleGeocodeEndpoint = "https://maps.googleapis.com/maps/api/geocode/json"

type googleGeocodeResponse struct {
	Status       string `json:"status"`
	ErrorMessage string `json:"error_message"`
	Results      []struct {
		Geometry struct {
			Location struct {
				Lat float64 `json:"lat"`
				Lng float64 `json:"lng"`
			} `json:"location"`
		} `json:"geometry"`
	} `json:"results"`
}

// googleGeocodeAddress resolves a free-text address to WGS84 coordinates using the Google Geocoding API.
func googleGeocodeAddress(ctx context.Context, apiKey, address string) (lat, lng float64, err error) {
	key := strings.TrimSpace(apiKey)
	addr := strings.TrimSpace(address)
	if key == "" {
		return 0, 0, fmt.Errorf("google geocoding api key is not configured (p_lacerate.googleGeocoding.apiKey)")
	}
	if addr == "" {
		return 0, 0, fmt.Errorf("address is empty")
	}
	u, err := url.Parse(googleGeocodeEndpoint)
	if err != nil {
		return 0, 0, err
	}
	q := u.Query()
	q.Set("address", addr)
	q.Set("key", key)
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return 0, 0, err
	}
	client := &http.Client{Timeout: 20 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return 0, 0, fmt.Errorf("geocoding request: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
	if err != nil {
		return 0, 0, err
	}
	var gr googleGeocodeResponse
	if err := json.Unmarshal(body, &gr); err != nil {
		slog.Error("lacerate: google geocode json parse", "error", err)
		return 0, 0, fmt.Errorf("geocoding response: invalid json")
	}
	switch gr.Status {
	case "OK":
		if len(gr.Results) == 0 {
			return 0, 0, fmt.Errorf("geocoding returned no results")
		}
		loc := gr.Results[0].Geometry.Location
		return loc.Lat, loc.Lng, nil
	case "ZERO_RESULTS":
		return 0, 0, fmt.Errorf("geocoding found no results for that address")
	default:
		msg := strings.TrimSpace(gr.ErrorMessage)
		if msg == "" {
			msg = gr.Status
		}
		return 0, 0, fmt.Errorf("geocoding failed: %s", msg)
	}
}
