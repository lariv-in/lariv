package p_pwa

import (
	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/lago"
	"maragu.dev/gomponents"
	"maragu.dev/gomponents/html"
)

type PwaIconConfig struct {
	Src   string `toml:"src" json:"src"`
	Sizes string `toml:"sizes" json:"sizes"`
	Type  string `toml:"type" json:"type,omitempty"`
}

type PwaAppleIconConfig struct {
	Src   string `toml:"src" json:"src"`
	Sizes string `toml:"sizes" json:"sizes"`
	Type  string `toml:"type" json:"type,omitempty"`
}

type PwaSplashScreenConfig struct {
	Src   string `toml:"src" json:"src"`
	Media string `toml:"media" json:"media"`
	Type  string `toml:"type" json:"type,omitempty"`
	Sizes string `toml:"sizes" json:"sizes,omitempty"`
}

type PwaShortcutConfig struct {
	Name        string `toml:"name" json:"name"`
	URL         string `toml:"url" json:"url"`
	Description string `toml:"description" json:"description,omitempty"`
}

type PwaScreenshotConfig struct {
	Src   string `toml:"src" json:"src"`
	Sizes string `toml:"sizes" json:"sizes,omitempty"`
	Type  string `toml:"type" json:"type,omitempty"`
}

// PwaConfig configures the endpoints served by this plugin:
// - /app.webmanifest
// - /serviceworker.js
// - /offline
// - /static/pwa/
type PwaConfig struct {
	// Optional filesystem path to a service worker JS file. If empty, a minimal
	// default service worker is served.
	ServiceWorkerPath string `toml:"serviceWorkerPath"`

	// Optional view key to serve for /offline. If empty, a minimal HTML page is served.
	OfflineViewName string `toml:"offlineViewName"`

	// Optional filesystem directory to serve under /static/pwa/.
	// If relative, it's resolved relative to the running binary's directory.
	// If empty, the route responds with 404.
	StaticDir string `toml:"staticDir"`

	// Manifest keys
	AppName                   string `toml:"PWA_APP_NAME"`
	AppDescription            string `toml:"PWA_APP_DESCRIPTION"`
	AppThemeColor             string `toml:"PWA_APP_THEME_COLOR"`
	AppBackgroundColor        string `toml:"PWA_APP_BACKGROUND_COLOR"`
	AppDisplay                string `toml:"PWA_APP_DISPLAY"`
	AppScope                  string `toml:"PWA_APP_SCOPE"`
	AppOrientation            string `toml:"PWA_APP_ORIENTATION"`
	AppStartURL               string `toml:"PWA_APP_START_URL"`
	AppPackageName            string `toml:"PWA_APP_PACKAGE_NAME"`
	AppSHA256CertFingerprints string `toml:"PWA_APP_SHA256_CERT_FINGERPRINTS"`

	AppStatusBarColor string `toml:"PWA_APP_STATUS_BAR_COLOR"`

	AppIcons        []PwaIconConfig         `toml:"PWA_APP_ICONS"`
	AppIconsApple   []PwaAppleIconConfig    `toml:"PWA_APP_ICONS_APPLE"`
	AppSplashScreen []PwaSplashScreenConfig `toml:"PWA_APP_SPLASH_SCREEN"`
	AppDir          string                  `toml:"PWA_APP_DIR"`
	AppLang         string                  `toml:"PWA_APP_LANG"`
	AppShortcuts    []PwaShortcutConfig     `toml:"PWA_APP_SHORTCUTS"`
	AppScreenshots  []PwaScreenshotConfig   `toml:"PWA_APP_SCREENSHOTS"`
}

var Config = &PwaConfig{}

func (c *PwaConfig) PostConfig() {
	if c.AppName != "" {
		err := components.RegistryShellHeadNodes.Register("base.title", html.TitleEl(gomponents.Text(c.AppName)))
		if err != nil {
			components.RegistryShellHeadNodes.Patch("base.title", func(_ gomponents.Node) gomponents.Node {
				return html.TitleEl(gomponents.Text(c.AppName))
			})
		}
	}
}

func init() {
	lago.RegistryConfig.Register("p_pwa", Config)
}
