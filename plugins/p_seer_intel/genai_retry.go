package p_seer_intel

import (
	"context"
	"errors"
	"log/slog"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"google.golang.org/genai"
)

const (
	genaiRetryMaxAttempts = 12
	genaiRetryBase        = 400 * time.Millisecond
	genaiRetryCap         = 60 * time.Second
	genaiRetryMaxShift    = 12
	// 429 / RESOURCE_EXHAUSTED: standard exponential backoff stayed under quota windows (see debug NDJSON: 8×429 in ~60s).
	genaiRetry429MinWait = 10 * time.Second
	genaiRetry429Cap     = 2 * time.Minute
	genaiRetry429Mul     = 4
)

func genAIRetryable(err error) bool {
	if err == nil {
		return false
	}
	var api genai.APIError
	if errors.As(err, &api) {
		switch api.Code {
		case http.StatusTooManyRequests, http.StatusServiceUnavailable,
			http.StatusBadGateway, http.StatusGatewayTimeout:
			return true
		}
		st := strings.ToUpper(api.Status)
		if strings.Contains(st, "RESOURCE_EXHAUSTED") || strings.Contains(st, "UNAVAILABLE") ||
			strings.Contains(st, "DEADLINE_EXCEEDED") {
			return true
		}
	}
	msg := strings.ToLower(err.Error())
	if strings.Contains(msg, "resource exhausted") || strings.Contains(msg, "resourceexhausted") {
		return true
	}
	if strings.Contains(msg, "quota") && (strings.Contains(msg, "exceed") || strings.Contains(msg, "limit")) {
		return true
	}
	if strings.Contains(msg, "rate limit") || strings.Contains(msg, "ratelimit") {
		return true
	}
	return false
}

func genAIResourceExhausted(err error) bool {
	if err == nil {
		return false
	}
	var api genai.APIError
	if errors.As(err, &api) {
		if api.Code == http.StatusTooManyRequests {
			return true
		}
		if strings.Contains(strings.ToUpper(api.Status), "RESOURCE_EXHAUSTED") {
			return true
		}
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "resource exhausted") || strings.Contains(msg, "error 429")
}

func genAIBackoffDelay(attempt int) time.Duration {
	if attempt < 0 {
		attempt = 0
	}
	shift := attempt
	if shift > genaiRetryMaxShift {
		shift = genaiRetryMaxShift
	}
	mul := int64(1) << shift
	d := genaiRetryBase * time.Duration(mul)
	if d > genaiRetryCap || d <= 0 {
		d = genaiRetryCap
	}
	jcap := d / 4
	if jcap > 2*time.Second {
		jcap = 2 * time.Second
	}
	var jitter time.Duration
	if jcap > 0 {
		jitter = time.Duration(rand.Int63n(int64(jcap) + 1))
	}
	return d + jitter
}

// genAIRetryWait picks sleep before the next attempt. Resource-exhausted / 429 errors use much longer waits
// so per-minute and short quota windows can recover (generic backoff was too aggressive in practice).
func genAIRetryWait(attempt int, err error) time.Duration {
	w := genAIBackoffDelay(attempt)
	if !genAIResourceExhausted(err) {
		return w
	}
	long := w * genaiRetry429Mul
	if long < genaiRetry429MinWait {
		long = genaiRetry429MinWait
	}
	if long > genaiRetry429Cap {
		long = genaiRetry429Cap
	}
	jcap := long / 8
	if jcap > 15*time.Second {
		jcap = 15 * time.Second
	}
	var jitter time.Duration
	if jcap > 0 {
		jitter = time.Duration(rand.Int63n(int64(jcap) + 1))
	}
	return long + jitter
}

// WithGenAIRetry runs fn until it succeeds, ctx is cancelled, attempts are exhausted, or err is not retryable.
// Waits truncated exponential backoff with additive jitter between attempts (quota / transient errors).
func WithGenAIRetry[T any](ctx context.Context, op string, fn func(context.Context) (T, error)) (T, error) {
	var zero T
	if genaiRetryMaxAttempts < 1 {
		return zero, errors.New("p_seer_intel: WithGenAIRetry: max attempts < 1")
	}
	for attempt := 0; attempt < genaiRetryMaxAttempts; attempt++ {
		v, err := fn(ctx)
		if err == nil {
			// #region agent log
			AgentDebugSessionLog("H2", "genai_retry.go:WithGenAIRetry", "genai_ok", map[string]any{
				"op": op, "attemptsUsed": attempt + 1,
			})
			// #endregion
			return v, nil
		}
		willRetry := attempt < genaiRetryMaxAttempts-1 && genAIRetryable(err)
		// #region agent log
		errSnippet := err.Error()
		if len(errSnippet) > 220 {
			errSnippet = errSnippet[:220] + "…"
		}
		AgentDebugSessionLog("H1", "genai_retry.go:WithGenAIRetry", "genai_attempt_failed", map[string]any{
			"op": op, "attempt": attempt + 1, "max": genaiRetryMaxAttempts, "willRetry": willRetry,
			"errSnippet": errSnippet,
		})
		// #endregion
		if attempt >= genaiRetryMaxAttempts-1 || !genAIRetryable(err) {
			return zero, err
		}
		wait := genAIRetryWait(attempt, err)
		// #region agent log
		AgentDebugSessionLog("H5", "genai_retry.go:WithGenAIRetry", "genai_backoff_wait", map[string]any{
			"op": op, "afterAttempt": attempt + 1, "waitMs": wait.Milliseconds(),
			"resourceExhausted": genAIResourceExhausted(err),
		})
		// #endregion
		slog.Warn("p_seer_intel: genai retry", "op", op, "attempt", attempt+1, "max", genaiRetryMaxAttempts, "wait", wait, "err", err)
		timer := time.NewTimer(wait)
		select {
		case <-ctx.Done():
			if !timer.Stop() {
				<-timer.C
			}
			return zero, ctx.Err()
		case <-timer.C:
		}
	}
	return zero, errors.New("p_seer_intel: WithGenAIRetry: unreachable")
}
