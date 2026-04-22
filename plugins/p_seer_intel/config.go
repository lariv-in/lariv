package p_seer_intel

import (
	"strings"

	"github.com/lariv-in/lago/lago"
)

// IntelConfig holds Intel-specific settings loaded from [Plugins.p_seer_intel].
type IntelConfig struct {
	GeocodingAPIKey string `toml:"geocodingApiKey"`
}

var IntelConfigValue = &IntelConfig{}

func (c *IntelConfig) PostConfig() {
	if c == nil {
		return
	}
	c.GeocodingAPIKey = strings.TrimSpace(c.GeocodingAPIKey)
}

func init() {
	lago.RegistryConfig.Register("p_seer_intel", IntelConfigValue)
}
