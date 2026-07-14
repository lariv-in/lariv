package p_llm_assistant

import (
	"strings"
)

type AssistantPluginConfig struct {
	CseAPIKey string `toml:"cseApiKey"`
	CseCX     string `toml:"cseCx"`
	ChatModel string `toml:"chatModel"`
}

const defaultAssistantChatModel = "gemini-2.5-flash"

var LlmAssistantPlugin = &AssistantPluginConfig{}

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

var AssistantAppConfig = struct {
	ChatMaxOutputTokens        int
	AssistantToolRounds        int
	IntelSearchLimitCap        int
	GoogleSearchResultLimitCap int
}{
	ChatMaxOutputTokens:        4096,
	AssistantToolRounds:        14,
	IntelSearchLimitCap:        20,
	GoogleSearchResultLimitCap: defaultGoogleSearchResultLimitCap,
}

func init() {
	registerPluginConfig("p_llm_assistant", LlmAssistantPlugin)
}
