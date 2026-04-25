package p_seer_opensky

import (
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
	openSkyTokenURL  = "https://auth.opensky-network.org/auth/realms/opensky-network/protocol/openid-connect/token"
	tokenRefreshSkew = 30 * time.Second
)

// openSkyTokenSource fetches and caches OAuth2 access tokens (client credentials).
type openSkyTokenSource struct {
	mu     sync.Mutex
	client *http.Client
	id     string
	secret string

	token     string
	expiresAt time.Time
}

func newOpenSkyTokenSource(c *http.Client, clientID, clientSecret string) *openSkyTokenSource {
	if c == nil {
		c = http.DefaultClient
	}
	return &openSkyTokenSource{client: c, id: clientID, secret: clientSecret}
}

// Token returns a valid Bearer token, refreshing when needed.
func (t *openSkyTokenSource) Token(ctx context.Context) (string, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	now := time.Now()
	if t.token != "" && now.Before(t.expiresAt) {
		return t.token, nil
	}
	return t.refreshLocked(ctx, now)
}

// Invalidate forces the next [Token] call to refresh.
func (t *openSkyTokenSource) Invalidate() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.token = ""
	t.expiresAt = time.Time{}
}

func (t *openSkyTokenSource) refreshLocked(ctx context.Context, now time.Time) (string, error) {
	if t.id == "" || t.secret == "" {
		return "", fmt.Errorf("p_seer_opensky: missing clientId or clientSecret")
	}
	body := url.Values{
		"grant_type":    {"client_credentials"},
		"client_id":     {t.id},
		"client_secret": {t.secret},
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, openSkyTokenURL, strings.NewReader(body.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := t.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("p_seer_opensky: token HTTP %d: %s", resp.StatusCode, truncateForLog(b))
	}
	var tr struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
	}
	if err := json.Unmarshal(b, &tr); err != nil {
		return "", fmt.Errorf("p_seer_opensky: token json: %w", err)
	}
	if tr.AccessToken == "" {
		return "", fmt.Errorf("p_seer_opensky: empty access_token")
	}
	exp := 30 * time.Minute
	if tr.ExpiresIn > 0 {
		exp = time.Duration(tr.ExpiresIn) * time.Second
	}
	t.token = tr.AccessToken
	t.expiresAt = now.Add(exp - tokenRefreshSkew)
	return t.token, nil
}

func truncateForLog(b []byte) string {
	const max = 200
	if len(b) <= max {
		return string(b)
	}
	return string(b[:max]) + "…"
}
