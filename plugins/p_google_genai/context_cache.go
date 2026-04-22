package p_google_genai

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"log/slog"
	"strings"
	"sync"
	"time"

	"google.golang.org/genai"
)

const defaultContextCacheTTLSeconds = 3600

type explicitContextCacheEntry struct {
	resourceName string
	expires      time.Time
}

var (
	explicitContextCacheMu sync.Mutex
	explicitContextByKey   = map[string]*explicitContextCacheEntry{}
)

func cacheKeyForSystem(model, systemPlain string) string {
	h := sha256.Sum256([]byte(model + "\n" + systemPlain))
	return hex.EncodeToString(h[:])
}

// attachExplicitContextCache optionally replaces inline SystemInstruction with a
// Gemini explicit context cache ([genai.Caches.Create]) when enabled in config.
// On success, cfg.CachedContent is set and cfg.SystemInstruction is cleared so
// the cached copy is not duplicated on the wire.
func attachExplicitContextCache(ctx context.Context, cli *genai.Client, model string, cfg *genai.GenerateContentConfig) {
	if cfg == nil || !GoogleGenAIConfig.ContextCacheEnabled {
		return
	}
	si := cfg.SystemInstruction
	if si == nil {
		return
	}
	sysPlain := strings.TrimSpace(joinContentText(si))
	if sysPlain == "" {
		return
	}
	key := cacheKeyForSystem(model, sysPlain)

	explicitContextCacheMu.Lock()
	defer explicitContextCacheMu.Unlock()

	ent := explicitContextByKey[key]
	if ent != nil && time.Now().Before(ent.expires.Add(-2*time.Minute)) {
		cfg.CachedContent = ent.resourceName
		cfg.SystemInstruction = nil
		return
	}

	ttl := time.Duration(GoogleGenAIConfig.ContextCacheTTLSeconds) * time.Second
	if ttl <= 0 {
		ttl = defaultContextCacheTTLSeconds * time.Second
	}

	cc, err := withGenAIRetryResp(ctx, "caches.create", func() (*genai.CachedContent, error) {
		return cli.Caches.Create(ctx, model, &genai.CreateCachedContentConfig{
			TTL:               ttl,
			SystemInstruction: si,
		})
	})
	if err != nil {
		slog.WarnContext(ctx, "p_google_genai: explicit context cache create failed; using uncached system instruction",
			"error", err.Error(), "model", model)
		return
	}
	if cc == nil || cc.Name == "" {
		slog.WarnContext(ctx, "p_google_genai: explicit context cache create returned empty name")
		return
	}
	exp := cc.ExpireTime
	if exp.IsZero() {
		exp = time.Now().Add(ttl)
	}
	explicitContextByKey[key] = &explicitContextCacheEntry{
		resourceName: cc.Name,
		expires:      exp,
	}
	cfg.CachedContent = cc.Name
	cfg.SystemInstruction = nil
}
