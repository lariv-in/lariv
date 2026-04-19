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
	genaiRetryMaxAttempts = 8
	genaiRetryBase        = 400 * time.Millisecond
	genaiRetryCap         = 60 * time.Second
	genaiRetryMaxShift    = 12
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
			return v, nil
		}
		if attempt >= genaiRetryMaxAttempts-1 || !genAIRetryable(err) {
			return zero, err
		}
		wait := genAIBackoffDelay(attempt)
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
