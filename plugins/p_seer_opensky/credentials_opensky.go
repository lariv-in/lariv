package p_seer_opensky

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// openskyCredsJSON matches a small JSON file for OAuth2 client credentials.
// Supported keys: client_id / client_secret (preferred) or clientId / clientSecret.
type openskyCredsJSON struct {
	ClientID        string `json:"client_id"`
	ClientSecret    string `json:"client_secret"`
	ClientIDAlt     string `json:"clientId"`
	ClientSecretAlt string `json:"clientSecret"`
}

func loadOpenSkyCredentialsFile(path string) (clientID, clientSecret string, err error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return "", "", fmt.Errorf("empty path")
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return "", "", err
	}
	var j openskyCredsJSON
	if err := json.Unmarshal(b, &j); err != nil {
		return "", "", err
	}
	id := strings.TrimSpace(j.ClientID)
	if id == "" {
		id = strings.TrimSpace(j.ClientIDAlt)
	}
	sec := strings.TrimSpace(j.ClientSecret)
	if sec == "" {
		sec = strings.TrimSpace(j.ClientSecretAlt)
	}
	if id == "" || sec == "" {
		return "", "", fmt.Errorf("JSON must include client_id and client_secret (or clientId and clientSecret)")
	}
	return id, sec, nil
}
