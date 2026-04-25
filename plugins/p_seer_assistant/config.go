package p_seer_assistant

import (
	"strings"

	"github.com/lariv-in/lago/lago"
)

// AssistantPluginConfig is loaded from TOML [Plugins.p_seer_assistant] when the app
// decodes the Plugins table (registry key "p_seer_assistant").
// Assistant chat uses [p_google_genai.NewClient] for the API key; requests use [google.golang.org/genai].
//
// Optional Google Custom Search JSON API credentials for the google_search tool:
//
//	cseApiKey = "..."
//	cseCx = "search-engine-id"
type AssistantPluginConfig struct {
	CseAPIKey string `toml:"cseApiKey"`
	CseCX     string `toml:"cseCx"`
	// ChatModel is the Gemini model id for assistant chat (e.g. "gemini-2.0-flash"). Empty uses default.
	ChatModel string `toml:"chatModel"`
}

const defaultAssistantChatModel = "gemini-2.5-flash"

var SeerAssistantPlugin = &AssistantPluginConfig{}

func (c *AssistantPluginConfig) PostConfig() {
	if c == nil {
		return
	}
	c.CseAPIKey = strings.TrimSpace(c.CseAPIKey)
	c.CseCX = strings.TrimSpace(c.CseCX)
	c.ChatModel = strings.TrimSpace(c.ChatModel)
	if c.ChatModel == "" {
		c.ChatModel = defaultAssistantChatModel
	}
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
