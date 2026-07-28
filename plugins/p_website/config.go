package p_website

import (
	"strings"

	"github.com/lariv-in/lariv"
	"github.com/lariv-in/lariv/registry"
)

// WebsiteConfig holds TOML settings for p_website.
type WebsiteConfig struct {
	// NewPageRootDir is the VNode directory path where "Create new HTML file"
	// places new page files (e.g. "website/pages"). Missing directories (and parents)
	// are created automatically. Empty means the filesystem root.
	NewPageRootDir string `toml:"newPageRootDir"`
	// AssetsDir is the VNode directory path where GrapesJS builder uploads are
	// stored (e.g. "website/assets"). Missing directories are created automatically.
	// Empty defaults to "{newPageRootDir}/assets" (or "assets" when newPageRootDir is empty).
	AssetsDir string `toml:"assetsDir"`
}

// Config is the active website plugin config (filled by Lariv PostConfig).
var Config = &WebsiteConfig{}

func (c *WebsiteConfig) PostConfig() {
	if c == nil {
		return
	}
	c.NewPageRootDir = strings.TrimSpace(c.NewPageRootDir)
	c.AssetsDir = strings.TrimSpace(c.AssetsDir)
}

// ResolvedAssetsDir returns the VNode directory used for builder asset uploads.
func (c *WebsiteConfig) ResolvedAssetsDir() string {
	if c == nil {
		return "assets"
	}
	if c.AssetsDir != "" {
		return c.AssetsDir
	}
	root := strings.Trim(c.NewPageRootDir, "/")
	if root == "" {
		return "assets"
	}
	return root + "/assets"
}

func pluginConfigs() lariv.PluginFeatures[lariv.Config] {
	return lariv.PluginFeatures[lariv.Config]{
		Entries: []registry.Pair[string, lariv.Config]{
			{Key: "p_website", Value: Config},
		},
	}
}
