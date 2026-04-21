package p_seer_gdelt

import "github.com/lariv-in/lago/lago"

type GDELTConfig struct {
	ProjectID         string `toml:"projectID"`
	CredentialsFile   string `toml:"credentialsFile"`
	Location          string `toml:"location"`
	DataProjectID     string `toml:"dataProjectID"`
	Dataset           string `toml:"dataset"`
	Table             string `toml:"table"`
	DefaultMaxRecords uint   `toml:"defaultMaxRecords"`
}

var Config = &GDELTConfig{}

func (c *GDELTConfig) PostConfig() {
	if c == nil {
		return
	}
	if c.Location == "" {
		c.Location = "US"
	}
	if c.DataProjectID == "" {
		c.DataProjectID = "gdelt-bq"
	}
	if c.Dataset == "" {
		c.Dataset = "gdeltv2"
	}
	if c.Table == "" {
		c.Table = "events"
	}
	if c.DefaultMaxRecords == 0 {
		c.DefaultMaxRecords = defaultGDELTMaxRecords
	}
	if c.DefaultMaxRecords > maxGDELTMaxRecords {
		c.DefaultMaxRecords = maxGDELTMaxRecords
	}
}

func init() {
	lago.RegistryConfig.Register("p_seer_gdelt", Config)
}
