package p_lacerate

import (
	"context"
	"log"
	"log/slog"
	"strings"

	"github.com/lariv-in/lago/lago"
)

// TwitterFetchMode selects how tweets are loaded for [TwitterSource].
type TwitterFetchMode string

const (
	TwitterFetchBearer   TwitterFetchMode = "bearer"
	TwitterFetchNitter   TwitterFetchMode = "nitter"
	TwitterFetchScraping TwitterFetchMode = "scraping"
)

// TwitterConfig holds global options for Twitter/X ingestion (see [lacerateConfig]).
type TwitterConfig struct {
	FetchMode     TwitterFetchMode `toml:"fetchMode"`
	BearerToken   string           `toml:"bearerToken"`
	NitterBaseURL string           `toml:"nitterBaseURL"`
}

// GeminiEmbeddingConfig selects the Gemini API embedding used by [VLEmbedder] ([GenAIVLEmbedder]).
// When APIKey is empty, no embedder is registered (embeddings stay zero / unset except where hooks clear them).
type GeminiEmbeddingConfig struct {
	APIKey string `toml:"apiKey"`
	Model  string `toml:"model"`
}

// GeminiAgentConfig selects the chat model for automated [Lookup] runs ([runLookupAgent]).
// Uses the same API key as [GeminiEmbeddingConfig]. When Model is empty, gemini-2.5-flash is used.
type GeminiAgentConfig struct {
	Model string `toml:"model"`
}

const (
	defaultIntelPreviewDirectory = "lacerate/intel_previews"
	// defaultIntelPreviewUserAgent is a normal browser UA so CDNs (e.g. external-preview.redd.it) accept preview fetches.
	defaultIntelPreviewUserAgent = "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36"
)

// IntelPreviewConfig holds the VFS path for stored preview images and the User-Agent for outbound preview HTTP requests.
type IntelPreviewConfig struct {
	Directory string `toml:"directory"`
	UserAgent string `toml:"userAgent"`
}

// lacerateConfig is the p_lacerate section in totschool.toml.
type lacerateConfig struct {
	Twitter         TwitterConfig         `toml:"twitter"`
	GeminiEmbedding GeminiEmbeddingConfig `toml:"geminiEmbedding"`
	GeminiAgent     GeminiAgentConfig     `toml:"geminiAgent"`
	IntelPreview    IntelPreviewConfig    `toml:"intelPreview"`
}

// Config is the active lacerate plugin configuration.
var Config = &lacerateConfig{}

func normalizeNitterBaseURL(raw string) string {
	s := strings.TrimSpace(raw)
	s = strings.TrimSuffix(s, "/")
	return s
}

func (c *lacerateConfig) PostConfig() {
	if strings.TrimSpace(c.IntelPreview.Directory) == "" {
		c.IntelPreview.Directory = defaultIntelPreviewDirectory
	}
	if strings.TrimSpace(c.IntelPreview.UserAgent) == "" {
		c.IntelPreview.UserAgent = defaultIntelPreviewUserAgent
	}

	c.Twitter.NitterBaseURL = normalizeNitterBaseURL(c.Twitter.NitterBaseURL)

	mode := strings.ToLower(strings.TrimSpace(string(c.Twitter.FetchMode)))
	switch TwitterFetchMode(mode) {
	case TwitterFetchBearer, TwitterFetchNitter, TwitterFetchScraping:
		c.Twitter.FetchMode = TwitterFetchMode(mode)
	default:
		slog.Error(`p_lacerate.twitter.fetchMode must be one of "bearer", "nitter", "scraping"`, "got", c.Twitter.FetchMode)
		log.Panicf(`p_lacerate.twitter.fetchMode must be one of "bearer", "nitter", "scraping" (got %q)`, c.Twitter.FetchMode)
	}

	switch c.Twitter.FetchMode {
	case TwitterFetchBearer:
		if strings.TrimSpace(c.Twitter.BearerToken) == "" {
			slog.Error(`p_lacerate.twitter.fetchMode is "bearer" but twitter.bearerToken is empty`)
			log.Panic(`p_lacerate.twitter.fetchMode is "bearer" but twitter.bearerToken is empty`)
		}
	case TwitterFetchNitter, TwitterFetchScraping:
		if c.Twitter.NitterBaseURL == "" {
			slog.Error("p_lacerate.twitter.nitterBaseURL is empty for fetch mode", "fetchMode", c.Twitter.FetchMode)
			log.Panicf(`p_lacerate.twitter.fetchMode is %q but twitter.nitterBaseURL is empty (use a Nitter instance base URL, no trailing slash)`, c.Twitter.FetchMode)
		}
	}

	key := strings.TrimSpace(c.GeminiEmbedding.APIKey)
	if key == "" {
		RegisterVLEmbedder(nil)
		return
	}
	emb, err := NewGenAIVLEmbedder(context.Background(), key, strings.TrimSpace(c.GeminiEmbedding.Model))
	if err != nil {
		slog.Error("p_lacerate: gemini embedding client", "error", err)
		log.Panicf("p_lacerate: gemini embedding client: %v", err)
	}
	RegisterVLEmbedder(emb)
}

func init() {
	lago.RegistryConfig.Register("p_lacerate", Config)
}
