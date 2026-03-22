package p_nirmancampus_website

import "github.com/lariv-in/lago/lago"

type WebsiteConfig struct {
	// Optional filesystem directory to serve under /nirman/static/.
	// If relative, it is resolved relative to the running binary's directory.
	// If empty, the route responds with 404.
	StaticDir string `toml:"staticDir"`
}

var Config = &WebsiteConfig{}

func (c *WebsiteConfig) PostConfig() {}

func init() {
	lago.RegistryConfig.Register("p_nirmancampus_website", Config)
}
