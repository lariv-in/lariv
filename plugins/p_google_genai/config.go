package p_google_genai

import (
	"strings"

	"github.com/lariv-in/lago/lago"
)

const (
	defaultTextModel        = "gemini-2.0-flash"
	defaultEmbeddingModel   = "text-embedding-004"
	defaultEmbeddingDimHint = 3072
	defaultTemperature      = float64(0.2)
	defaultTopP             = float64(0.9)
	defaultTopK             = float32(40)
	defaultThinkingMode     = ThinkingModeDisabled
	defaultBackend          = "gemini"
	defaultMaxOutputTokens  = 8192
)

const (
	ThinkingModeDefault  = "default"
	ThinkingModeDisabled = "disabled"
	ThinkingModeEnabled  = "enabled"
)

// BackendGemini uses the Gemini Developer API (API key).
// BackendVertex uses Vertex AI (project, location, ADC).
const (
	BackendGemini = "gemini"
	BackendVertex = "vertex"
)

// Config is loaded only when the app config file defines a matching Plugins table:
//
//	[Plugins.p_google_genai]
//	apiKey = "..."     # or api_key = "..." (snake_case alias)
//	contextCacheEnabled = false   # optional: explicit context cache for large system prompts (Gemini Caches API)
//	contextCacheTTLSeconds = 3600
//
// The registry key must be exactly "p_google_genai" (see [lago.LoadConfigFromFile]).
// If that block is missing, nothing is decoded and APIKey stays empty; the GenAI SDK
// then falls back to GOOGLE_API_KEY / GEMINI_API_KEY.
type Config struct {
	Backend          string  `toml:"backend"`
	APIKey           string  `toml:"apiKey"`
	APIKeySnake      string  `toml:"api_key"`
	Project          string  `toml:"project"`
	Location         string  `toml:"location"`
	TextModel        string  `toml:"textModel"`
	EmbeddingModel   string  `toml:"embeddingModel"`
	TextTemperature  float64 `toml:"textTemperature"`
	TextTopK         float32 `toml:"textTopK"`
	TextTopP         float64 `toml:"textTopP"`
	ThinkingMode     string  `toml:"thinkingMode"`
	ThinkingBudget   int     `toml:"thinkingBudget"`
	MaxOutputTokens  int     `toml:"maxOutputTokens"`
	EmbeddingDimHint int     `toml:"embeddingDimHint"`
	// RetryMax is extra attempts after the first call for 429 / RESOURCE_EXHAUSTED and similar (max 15; total tries = 1+RetryMax, default 10).
	RetryMax int `toml:"retryMax"`
	// RetryBaseMillis is initial backoff; delays grow exponentially with jitter (min 50).
	RetryBaseMillis int `toml:"retryBaseMillis"`

	// ContextCacheEnabled uses Gemini explicit context caching ([google.golang.org/genai.Caches]):
	// system instructions are stored server-side and referenced per request (lower latency / cost for large prompts).
	// Requires sufficient cached token minimums per Google’s API; creation failures fall back to uncached requests.
	ContextCacheEnabled bool `toml:"contextCacheEnabled"`
	// ContextCacheTTLSeconds is TTL for newly created cache entries (default 3600 when enabled).
	ContextCacheTTLSeconds int `toml:"contextCacheTTLSeconds"`
}

var GoogleGenAIConfig = &Config{}

func (c *Config) PostConfig() {
	if c == nil {
		return
	}
	c.Backend = strings.ToLower(strings.TrimSpace(c.Backend))
	c.APIKey = strings.TrimSpace(c.APIKey)
	c.APIKeySnake = strings.TrimSpace(c.APIKeySnake)
	if c.APIKey == "" {
		c.APIKey = c.APIKeySnake
	}
	c.APIKeySnake = ""
	c.Project = strings.TrimSpace(c.Project)
	c.Location = strings.TrimSpace(c.Location)
	c.TextModel = strings.TrimSpace(c.TextModel)
	c.EmbeddingModel = strings.TrimSpace(c.EmbeddingModel)
	if c.Backend == "" {
		c.Backend = defaultBackend
	}
	if c.TextModel == "" {
		c.TextModel = defaultTextModel
	}
	if c.EmbeddingModel == "" {
		c.EmbeddingModel = defaultEmbeddingModel
	}
	if c.TextTemperature < 0 {
		c.TextTemperature = defaultTemperature
	}
	if c.TextTopP <= 0 || c.TextTopP > 1 {
		c.TextTopP = defaultTopP
	}
	if c.TextTopK <= 0 {
		c.TextTopK = defaultTopK
	}
	switch strings.ToLower(strings.TrimSpace(c.ThinkingMode)) {
	case "", ThinkingModeDefault:
		c.ThinkingMode = defaultThinkingMode
	case ThinkingModeDisabled, ThinkingModeEnabled:
		c.ThinkingMode = strings.ToLower(strings.TrimSpace(c.ThinkingMode))
	default:
		c.ThinkingMode = defaultThinkingMode
	}
	if c.ThinkingBudget < 0 {
		c.ThinkingBudget = 0
	}
	if c.MaxOutputTokens < 0 {
		c.MaxOutputTokens = 0
	}
	if c.EmbeddingDimHint <= 0 {
		c.EmbeddingDimHint = defaultEmbeddingDimHint
	}
	if c.RetryMax < 0 {
		c.RetryMax = defaultRetryMax
	}
	if c.RetryMax > maxRetryAttemptsCap {
		c.RetryMax = maxRetryAttemptsCap
	}
	if c.RetryBaseMillis <= 0 {
		c.RetryBaseMillis = defaultRetryBaseMillis
	}
	if c.RetryBaseMillis < 50 {
		c.RetryBaseMillis = 50
	}
	if c.ContextCacheTTLSeconds < 0 {
		c.ContextCacheTTLSeconds = 0
	}
}

func init() {
	GoogleGenAIConfig.TextTemperature = defaultTemperature
	GoogleGenAIConfig.TextTopP = defaultTopP
	GoogleGenAIConfig.TextTopK = defaultTopK
	GoogleGenAIConfig.ThinkingMode = defaultThinkingMode
	GoogleGenAIConfig.TextModel = defaultTextModel
	GoogleGenAIConfig.EmbeddingModel = defaultEmbeddingModel
	GoogleGenAIConfig.Backend = defaultBackend
	GoogleGenAIConfig.MaxOutputTokens = defaultMaxOutputTokens
	GoogleGenAIConfig.RetryMax = defaultRetryMax
	GoogleGenAIConfig.RetryBaseMillis = defaultRetryBaseMillis
	lago.RegistryConfig.Register("p_google_genai", GoogleGenAIConfig)
}
