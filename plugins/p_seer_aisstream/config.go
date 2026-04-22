package p_seer_aisstream

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"sync/atomic"

	"github.com/lariv-in/lago/lago"
)

// AISStreamConfig holds API key for aisstream.io (server-side WebSocket only).
// Set apiKey in TOML, or use apiKeyFile for JSON: {"apiKey": "..."}.
type AISStreamConfig struct {
	APIKey     string `toml:"apiKey"`
	APIKeyFile string `toml:"apiKeyFile"`
}

var Config = &AISStreamConfig{}

var configAPIKey atomic.Value // string, resolved key after PostConfig

func (c *AISStreamConfig) PostConfig() {
	if c == nil {
		configAPIKey.Store("")
		return
	}
	key := strings.TrimSpace(c.APIKey)
	if key == "" && c.APIKeyFile != "" {
		loaded, err := loadAPIKeyFile(c.APIKeyFile)
		if err != nil {
			slog.Warn("p_seer_aisstream: could not load api key file", "path", c.APIKeyFile, "error", err)
		} else {
			key = loaded
		}
	}
	configAPIKey.Store(key)
}

// EffectiveAPIKey returns the configured key after TOML load (file merged in PostConfig).
func EffectiveAPIKey() string {
	if v := configAPIKey.Load(); v != nil {
		if s, ok := v.(string); ok && s != "" {
			return s
		}
	}
	return strings.TrimSpace(Config.APIKey)
}

func loadAPIKeyFile(path string) (string, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return "", fmt.Errorf("empty path")
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil {
		return "", err
	}
	var k string
	if v, ok := m["apiKey"].(string); ok {
		k = v
	}
	if k == "" {
		if v, ok := m["APIKey"].(string); ok {
			k = v
		}
	}
	k = strings.TrimSpace(k)
	if k == "" {
		return "", fmt.Errorf("JSON must include apiKey or APIKey")
	}
	return k, nil
}

func init() {
	lago.RegistryConfig.Register("p_seer_aisstream", Config)
}
