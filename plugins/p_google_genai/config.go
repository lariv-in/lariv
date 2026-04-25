package p_google_genai

import (
	"strings"

	"github.com/lariv-in/lago/lago"
)

type Config struct {
	APIKey string `toml:"apiKey"`
}

var GoogleGenAIConfig = &Config{}

func (c *Config) PostConfig() {
	if c == nil {
		return
	}
	c.APIKey = strings.TrimSpace(c.APIKey)
}

func init() {
	lago.RegistryConfig.Register("p_google_genai", GoogleGenAIConfig)
}
