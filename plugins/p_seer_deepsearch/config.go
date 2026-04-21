package p_seer_deepsearch

import "github.com/lariv-in/lago/lago"

// DeepSearchConfig holds Google Programmable Search Engine (Custom Search JSON API) settings.
// Loaded from TOML under [Plugins.p_seer_deepsearch].
//
//	[Plugins.p_seer_deepsearch]
//	apiKey = "…"
//	cx = "your-search-engine-id"
//	reportMaxOutputTokens = 16384   # optional; Gemini generateContent max output (default 8192, cap 65536)
//	expandMaxOutputTokens = 2048    # optional; query-expansion LLM (default 1024, cap 8192)
//
// Gemini API key and models come from [Plugins.p_seer_intel] ([p_seer_intel.IntelGenAI]).
type DeepSearchConfig struct {
	APIKey string `toml:"apiKey"`
	CX     string `toml:"cx"`
	// ReportMaxOutputTokens caps the final markdown report LLM (GenerateContent maxOutputTokens). Zero → 8192 in [PostConfig].
	ReportMaxOutputTokens int `toml:"reportMaxOutputTokens"`
	// ExpandMaxOutputTokens caps the query-expansion LLM. Zero → 1024 in [PostConfig].
	ExpandMaxOutputTokens int `toml:"expandMaxOutputTokens"`
}

// DeepSearchAppConfig is the package-level config (filled after config load).
var DeepSearchAppConfig = &DeepSearchConfig{}

const (
	defaultDeepSearchReportMaxOutputTokens = 8192
	maxDeepSearchReportMaxOutputTokens     = 65536
	defaultDeepSearchExpandMaxOutputTokens = 1024
	maxDeepSearchExpandMaxOutputTokens     = 8192
)

func (c *DeepSearchConfig) PostConfig() {
	if c == nil {
		return
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
