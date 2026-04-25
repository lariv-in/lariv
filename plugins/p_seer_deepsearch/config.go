package p_seer_deepsearch

import (
	"strings"

	"github.com/lariv-in/lago/lago"
)

// DeepSearchConfig holds Google Programmable Search Engine (Custom Search JSON API) settings.
// Loaded from TOML under [Plugins.p_seer_deepsearch].
//
//	[Plugins.p_seer_deepsearch]
//	apiKey = "…"
//	cx = "your-search-engine-id"
//	reportMaxOutputTokens = 16384   # optional; final report max output tokens (default 8192, cap 65536)
//	expandMaxOutputTokens = 2048    # optional; query-expansion max output tokens (default 1024, cap 8192)
//
// LlmModel selects Gemini for expand / tool / report; API key comes from [Plugins.p_google_genai].
type DeepSearchConfig struct {
	APIKey string `toml:"apiKey"`
	CX     string `toml:"cx"`
	// LlmModel is the Gemini model id (e.g. "gemini-2.0-flash"). Empty → default in [PostConfig].
	LlmModel string `toml:"llmModel"`
	// ReportMaxOutputTokens caps the final markdown report output. Zero → 8192 in [PostConfig].
	ReportMaxOutputTokens int `toml:"reportMaxOutputTokens"`
	// ExpandMaxOutputTokens caps query-expansion output. Zero → 1024 in [PostConfig].
	ExpandMaxOutputTokens int `toml:"expandMaxOutputTokens"`
}

// DeepSearchAppConfig is the package-level config (filled after config load).
var DeepSearchAppConfig = &DeepSearchConfig{}

const (
	defaultDeepSearchReportMaxOutputTokens = 8192
	maxDeepSearchReportMaxOutputTokens     = 65536
	defaultDeepSearchExpandMaxOutputTokens = 1024
	maxDeepSearchExpandMaxOutputTokens     = 8192
	defaultDeepSearchLlmModel              = "gemini-2.0-flash"
)

func (c *DeepSearchConfig) PostConfig() {
	if c == nil {
		return
	}
	c.LlmModel = strings.TrimSpace(c.LlmModel)
	if c.LlmModel == "" {
		c.LlmModel = defaultDeepSearchLlmModel
	}
	if c.ReportMaxOutputTokens <= 0 {
		c.ReportMaxOutputTokens = defaultDeepSearchReportMaxOutputTokens
	}
	if c.ReportMaxOutputTokens > maxDeepSearchReportMaxOutputTokens {
		c.ReportMaxOutputTokens = maxDeepSearchReportMaxOutputTokens
	}
	if c.ExpandMaxOutputTokens <= 0 {
		c.ExpandMaxOutputTokens = defaultDeepSearchExpandMaxOutputTokens
	}
	if c.ExpandMaxOutputTokens > maxDeepSearchExpandMaxOutputTokens {
		c.ExpandMaxOutputTokens = maxDeepSearchExpandMaxOutputTokens
	}
}

func init() {
	lago.RegistryConfig.Register("p_seer_deepsearch", DeepSearchAppConfig)
}
