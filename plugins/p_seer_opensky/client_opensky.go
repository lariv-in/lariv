package p_seer_opensky

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

const (
	openskyStatesURL = "https://opensky-network.org/api/states/all"
	openskyTokenURL  = "https://auth.opensky-network.org/auth/realms/opensky-network/protocol/openid-connect/token"
	tokenRefreshSlack = 30 * time.Second
	upstreamTimeout   = 25 * time.Second
)

// Aircraft is normalized state for JSON responses and tests.
type Aircraft struct {
	Icao24   string   `json:"icao24"`
	Lat      float64  `json:"lat"`
	Lng      float64  `json:"lng"`
	Heading  *float64 `json:"heading,omitempty"`
	Velocity *float64 `json:"velocity,omitempty"`
	Callsign *string  `json:"callsign,omitempty"`
	OnGround bool     `json:"onGround"`
	Altitude *float64 `json:"altitude,omitempty"`
}

// StatesResponse is our public JSON shape for GET /api/states/.
type StatesResponse struct {
	Time     int64      `json:"time"`
	Aircraft []Aircraft `json:"aircraft"`
}

type rawStatesBody struct {
	Time   int64    `json:"time"`
	States [][]any  `json:"states"`
}

var (
	httpClient = &http.Client{Timeout: upstreamTimeout}

	tokenMu      sync.Mutex
	cachedToken  string
	tokenExpires time.Time
)

// ParseOpenSkyStatesBody parses an OpenSky /states/all JSON body into normalized aircraft.
func ParseOpenSkyStatesBody(body []byte) (StatesResponse, error) {
	var raw rawStatesBody
	if err := json.Unmarshal(body, &raw); err != nil {
		return StatesResponse{}, err
	}
	out := StatesResponse{Time: raw.Time}
	if raw.States == nil {
		return out, nil
	}
	for _, row := range raw.States {
		if a, ok := normalizeStateRow(row); ok {
			out.Aircraft = append(out.Aircraft, a)
		}
	}
	return out, nil
}

func normalizeStateRow(row []any) (Aircraft, bool) {
	if len(row) < 11 {
		return Aircraft{}, false
	}
	icao, ok := row[0].(string)
	if !ok || icao == "" {
		return Aircraft{}, false
	}
	lon, ok1 := anyFloat(row[5])
	lat, ok2 := anyFloat(row[6])
	if !ok1 || !ok2 {
		return Aircraft{}, false
	}
	var alt *float64
	if v, ok := anyFloat(row[7]); ok {
		alt = &v
	}
	onGround := false
	switch v := row[8].(type) {
	case bool:
		onGround = v
	}

	var vel *float64
	if v, ok := anyFloat(row[9]); ok {
		vel = &v
	}
	var heading *float64
	if v, ok := anyFloat(row[10]); ok {
		heading = &v
	}

	var callsign *string
	if row[1] != nil {
		if s, ok := row[1].(string); ok {
			s = strings.TrimSpace(s)
			if s != "" {
				callsign = &s
			}
		}
	}

	return Aircraft{
		Icao24:   strings.ToLower(icao),
		Lat:      lat,
		Lng:      lon,
		Heading:  heading,
		Velocity: vel,
		Callsign: callsign,
		OnGround: onGround,
		Altitude: alt,
	}, true
}

func anyFloat(v any) (float64, bool) {
	switch x := v.(type) {
	case float64:
		return x, true
	case json.Number:
		f, err := x.Float64()
		return f, err == nil
	default:
		return 0, false
	}
}

func (c *OpenSkyConfig) bearerToken(ctx context.Context) (string, error) {
	if c == nil || c.ClientID == "" || c.ClientSecret == "" {
		return "", nil
	}
	tokenMu.Lock()
	defer tokenMu.Unlock()
	if cachedToken != "" && time.Until(tokenExpires) > tokenRefreshSlack {
		return cachedToken, nil
	}
	form := url.Values{}
	form.Set("grant_type", "client_credentials")
	form.Set("client_id", c.ClientID)
	form.Set("client_secret", c.ClientSecret)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, openskyTokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	res, err := httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return "", err
	}
	if res.StatusCode != http.StatusOK {
		return "", fmt.Errorf("opensky token: status %d: %s", res.StatusCode, bytes.TrimSpace(body))
	}
	var tok struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
	}
	if err := json.Unmarshal(body, &tok); err != nil {
		return "", err
	}
	if tok.AccessToken == "" {
		return "", fmt.Errorf("opensky token: empty access_token")
	}
	exp := time.Duration(tok.ExpiresIn) * time.Second
	if exp <= 0 {
		exp = 30 * time.Minute
	}
	cachedToken = tok.AccessToken
	tokenExpires = time.Now().Add(exp - tokenRefreshSlack)
	return cachedToken, nil
}

// FetchStates calls OpenSky /states/all with the given raw query string (e.g. "lamin=..&lomin=..").
func FetchStates(ctx context.Context, rawQuery string) (StatesResponse, int, http.Header, error) {
	u := openskyStatesURL
	if rawQuery != "" {
		u += "?" + rawQuery
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return StatesResponse{}, 0, nil, err
	}
	tok, err := Config.bearerToken(ctx)
	if err != nil {
		return StatesResponse{}, 0, nil, err
	}
	if tok != "" {
		req.Header.Set("Authorization", "Bearer "+tok)
	}
	res, err := httpClient.Do(req)
	if err != nil {
		return StatesResponse{}, 0, nil, err
	}
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return StatesResponse{}, res.StatusCode, res.Header, err
	}
	if res.StatusCode != http.StatusOK {
		return StatesResponse{}, res.StatusCode, res.Header, fmt.Errorf("opensky: status %d: %s", res.StatusCode, bytes.TrimSpace(body))
	}
	parsed, err := ParseOpenSkyStatesBody(body)
	if err != nil {
		return StatesResponse{}, res.StatusCode, res.Header, err
	}
	return parsed, res.StatusCode, res.Header, nil
}
