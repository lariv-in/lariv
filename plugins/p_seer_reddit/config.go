package p_seer_reddit

import (
	"strings"

	"github.com/lariv-in/lago/lago"
)

// SeerRedditPlugin is TOML config for [p_seer_reddit] (e.g. Reddit source LLM filter).
type SeerRedditPlugin struct {
	// FilterLlmModel is the Gemini model id for filter gate (e.g. "gemini-2.5-flash"). Empty → default.
	FilterLlmModel string `toml:"filterLlmModel"`
	// FilterLlmMaxOutputTokens caps structured JSON; ≤0 → default.
	FilterLlmMaxOutputTokens int `toml:"filterLlmMaxOutputTokens"`
}

// RedditPlugin is the app config for this plugin. Registered in init.
var RedditPlugin = &SeerRedditPlugin{}

const defaultFilterLlmModel = "gemini-2.5-flash"

const defaultFilterLlmMaxOutputTokens int32 = 256

func (c *SeerRedditPlugin) PostConfig() {
	if c == nil {
		return
	}
	c.FilterLlmModel = strings.TrimSpace(c.FilterLlmModel)
	if c.FilterLlmModel == "" {
		c.FilterLlmModel = defaultFilterLlmModel
	}
}

func init() {
	lago.RegistryConfig.Register("p_seer_reddit", RedditPlugin)
}

func redditFilterLlmModel() string {
	if RedditPlugin == nil {
		return defaultFilterLlmModel
	}
	if m := strings.TrimSpace(RedditPlugin.FilterLlmModel); m != "" {
		return m
	}
	return defaultFilterLlmModel
}

func redditFilterLlmMaxOutputTokens() int32 {
	n := defaultFilterLlmMaxOutputTokens
	if RedditPlugin != nil && RedditPlugin.FilterLlmMaxOutputTokens > 0 {
		n = int32(RedditPlugin.FilterLlmMaxOutputTokens)
	}
	if n < 32 {
		n = 32
	}
	if n > 1024 {
		n = 1024
	}
	return n
}
