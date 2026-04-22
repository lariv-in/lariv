package p_seer_assistant

import (
	"strings"

	"github.com/lariv-in/lago/lago"
)

// AssistantPluginConfig is loaded from TOML [Plugins.p_seer_assistant] when the app
// decodes the Plugins table (registry key "p_seer_assistant").
// Assistant chat streaming uses github.com/lariv-in/lago/plugins/p_google_genai only.
//
// Optional Google Custom Search JSON API credentials for the google_search tool:
//
//	cseApiKey = "..."
//	cseCx = "search-engine-id"
type AssistantPluginConfig struct {
	CseAPIKey string `toml:"cseApiKey"`
	CseCX     string `toml:"cseCx"`
}

var SeerAssistantPlugin = &AssistantPluginConfig{}

func (c *AssistantPluginConfig) PostConfig() {
	if c == nil {
		return
	}
	c.CseAPIKey = strings.TrimSpace(c.CseAPIKey)
	c.CseCX = strings.TrimSpace(c.CseCX)
}

const (
	defaultGoogleSearchResultLimitCap = 20
	assistantGoogleSearchMaxPages     = 2
)

// AssistantAppConfig holds runtime tuning for the assistant WS handler.
var AssistantAppConfig = struct {
	ChatMaxOutputTokens int
	AssistantToolRounds int
	IntelSearchLimitCap int
	// GoogleSearchResultLimitCap caps results per google_search tool call (max 20 = 2 CSE pages of 10).
	GoogleSearchResultLimitCap int
}{
	ChatMaxOutputTokens:        1536,
	AssistantToolRounds:        14,
	IntelSearchLimitCap:        20,
	GoogleSearchResultLimitCap: defaultGoogleSearchResultLimitCap,
}

func init() {
	lago.RegistryConfig.Register("p_seer_assistant", SeerAssistantPlugin)
}
