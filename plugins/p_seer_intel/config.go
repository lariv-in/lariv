package p_seer_intel

import "github.com/lariv-in/lago/lago"

// IntelGenAIConfig holds Google Gen AI (Gemini Developer API) settings used by [NewFromIntelKind].
// Values are loaded from the app config file under [Plugins] key "p_seer_intel" (see [lago.LoadConfigFromFile]).
//
// Example TOML fragment:
//
//	[Plugins.p_seer_intel]
//	apiKey = "…"
//	llmModel = "gemini-2.0-flash"
//	embeddingModel = "gemini-embedding-2-preview"
type IntelGenAIConfig struct {
	APIKey          string `toml:"apiKey"`
	LLMModel        string `toml:"llmModel"`
	EmbeddingModel  string `toml:"embeddingModel"`
}

const (
	defaultIntelLLMModel       = "gemini-2.0-flash"
	defaultIntelEmbeddingModel = "gemini-embedding-2-preview"
)

// IntelGenAI is the package-level Gen AI configuration (filled from registry after config load).
var IntelGenAI = &IntelGenAIConfig{}

func (c *IntelGenAIConfig) PostConfig() {
	if c == nil {
		return
	}
	if c.LLMModel == "" {
		c.LLMModel = defaultIntelLLMModel
	}
	if c.EmbeddingModel == "" {
		c.EmbeddingModel = defaultIntelEmbeddingModel
	}
}

func init() {
	// Defaults apply even when no [Plugins.p_seer_intel] block exists; TOML decode then overrides.
	IntelGenAI.LLMModel = defaultIntelLLMModel
	IntelGenAI.EmbeddingModel = defaultIntelEmbeddingModel
	lago.RegistryConfig.Register("p_seer_intel", IntelGenAI)
}
