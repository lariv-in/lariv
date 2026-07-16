package p_google_genai

import (
	"strings"

	"github.com/lariv-in/lariv"
	"github.com/lariv-in/lariv/registry"
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

func pluginConfigs() lariv.PluginFeatures[lariv.Config] {
	return lariv.PluginFeatures[lariv.Config]{
		Entries: []registry.Pair[string, lariv.Config]{
			{Key: "p_google_genai", Value: GoogleGenAIConfig},
		},
	}
}
