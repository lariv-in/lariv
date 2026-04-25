package p_seer_opensky

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const openSkyStatesAllURL = "https://opensky-network.org/api/states/all"

// FetchAllStatesGET performs GET [openSkyStatesAllURL] with Bearer auth, decodes [StatesEnvelope].
func FetchAllStatesGET(ctx context.Context, client *http.Client, tok *openSkyTokenSource) (*StatesEnvelope, error) {
	if client == nil {
		client = http.DefaultClient
	}
	return fetchStatesOnce(ctx, client, tok, false)
}

func fetchStatesOnce(ctx context.Context, client *http.Client, tok *openSkyTokenSource, after401 bool) (*StatesEnvelope, error) {
	token, err := tok.Token(ctx)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, openSkyStatesAllURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode == http.StatusUnauthorized && !after401 {
		tok.Invalidate()
		return fetchStatesOnce(ctx, client, tok, true)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("p_seer_opensky: states HTTP %d: %s", resp.StatusCode, truncateForLog(body))
	}
	var env StatesEnvelope
	if err := json.Unmarshal(body, &env); err != nil {
		return nil, fmt.Errorf("p_seer_opensky: states json: %w", err)
	}
	return &env, nil
}

// newFetchHTTPClient returns a client for OpenSky with sane timeouts.
func newFetchHTTPClient() *http.Client {
	return &http.Client{
		Timeout: 90 * time.Second,
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
		},
	}
}
