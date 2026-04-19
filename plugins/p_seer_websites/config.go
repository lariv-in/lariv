package p_seer_websites

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/lariv-in/lago/lago"
)

// WebsiteRodConfig controls headless Chromium launch for website scraping ([getRodBrowser]).
// Loaded from the app config under [Plugins] key "p_seer_websites" (see [lago.LoadConfigFromFile]).
//
// Example TOML fragment:
//
//	[Plugins.p_seer_websites]
//	userDataDir = "/home/you/.config/chromium"
//	profileDir = "Default"
//
// [UserDataDir] is passed to Chrome as --user-data-dir. Default matches Chromium’s usual profile root
// (Linux: ~/.config/chromium, macOS: ~/Library/Application Support/Chromium, Windows: %LOCALAPPDATA%\Chromium\User Data).
// Only one process may use a given user-data-dir at a time (quit GUI Chromium before scraping if they match).
//
// [ProfileDir] is optional; when set, passed as --profile-directory (e.g. "Default", "Profile 1").
type WebsiteRodConfig struct {
	UserDataDir string `toml:"userDataDir"`
	ProfileDir  string `toml:"profileDir"`
}

// WebsiteRod is the package-level Rod/Chromium launch config (filled from registry after config load).
var WebsiteRod = &WebsiteRodConfig{}

func defaultChromiumUserDataDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	switch runtime.GOOS {
	case "darwin":
		return filepath.Join(home, "Library", "Application Support", "Chromium")
	case "windows":
		return filepath.Join(home, "AppData", "Local", "Chromium", "User Data")
	default:
		return filepath.Join(home, ".config", "chromium")
	}
}

func (c *WebsiteRodConfig) PostConfig() {
	if c == nil {
		return
	}
	if strings.TrimSpace(c.UserDataDir) == "" {
		c.UserDataDir = defaultChromiumUserDataDir()
	}
}

func init() {
	if d := defaultChromiumUserDataDir(); d != "" {
		WebsiteRod.UserDataDir = d
	}
	lago.RegistryConfig.Register("p_seer_websites", WebsiteRod)
}
