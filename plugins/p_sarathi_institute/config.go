package p_sarathi_institute

import "github.com/lariv-in/lago/lago"

// Config holds singleton institute/org metadata for Sarathi (no DB tables).
// Loaded from [Plugins.p_sarathi_institute] in the deployment TOML.
type Config struct {
	Name        string   `toml:"name"`
	ShortName   string   `toml:"short_name"`
	LegalName   string   `toml:"legal_name"`
	Email       string   `toml:"email"`
	Phone       string   `toml:"phone"`
	Address     string   `toml:"address"`
	Website     string   `toml:"website"`
	LogoPath    string   `toml:"logo_path"`
	Timezone    string   `toml:"timezone"`
	FeatureTags []string `toml:"feature_tags"`
}

func (c *Config) PostConfig() {}

// Singleton decoded at startup (same pointer RegistryConfig holds).
var Institute = &Config{}

func init() {
	lago.RegistryConfig.Register("p_sarathi_institute", Institute)
}
