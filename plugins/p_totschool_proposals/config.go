package p_totschool_proposals

import "github.com/lariv-in/lago/lago"

type AIConfig struct {
	APIKey string `toml:"apiKey"`
	Model  string `toml:"model"`
}

var aiConfig = &AIConfig{}

func (c *AIConfig) PostConfig() {}

func init() {
	lago.RegistryConfig.Register("p_totschool_proposals", aiConfig)
}
