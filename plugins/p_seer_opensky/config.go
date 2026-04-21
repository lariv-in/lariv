package p_seer_opensky

import (
	"log/slog"

	"github.com/lariv-in/lago/lago"
)

// OpenSkyConfig holds optional OAuth2 client credentials for OpenSky API
// (higher rate limits and finer time resolution). Leave empty for anonymous
// requests. See https://openskynetwork.github.io/opensky-api/rest.html
//
// Set clientID and clientSecret in TOML, or point credentialsFile at a JSON file
// with client_id / client_secret. TOML values override file when both are set.
type OpenSkyConfig struct {
	ClientID        string `toml:"clientID"`
	ClientSecret    string `toml:"clientSecret"`
	CredentialsFile string `toml:"credentialsFile"`
}

var Config = &OpenSkyConfig{}

func (c *OpenSkyConfig) PostConfig() {
	if c == nil || c.CredentialsFile == "" {
		return
	}
	fid, fsec, err := loadOpenSkyCredentialsFile(c.CredentialsFile)
	if err != nil {
		slog.Warn("p_seer_opensky: could not load credentials file", "path", c.CredentialsFile, "error", err)
		return
	}
	if c.ClientID == "" {
		c.ClientID = fid
	}
	if c.ClientSecret == "" {
		c.ClientSecret = fsec
	}
}

func init() {
	lago.RegistryConfig.Register("p_seer_opensky", Config)
}
