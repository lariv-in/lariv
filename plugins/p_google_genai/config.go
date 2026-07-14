package p_google_genai

import (
	"strings"

	"github.com/lariv-in/lago"
	"github.com/lariv-in/lago/registry"
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

func pluginConfigs() lago.PluginFeatures[lago.Config] {
	return lago.PluginFeatures[lago.Config]{
		Entries: []registry.Pair[string, lago.Config]{
			{Key: "p_google_genai", Value: GoogleGenAIConfig},
		},
	}
}
