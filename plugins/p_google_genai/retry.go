package p_google_genai

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"math/rand/v2"
	"net/http"
	"strings"
	"time"

	"google.golang.org/genai"
)

const (
	defaultMaxAttempts     = 10
	defaultRetryMax        = defaultMaxAttempts - 1 // extra tries after first; total = defaultMaxAttempts
	defaultRetryBaseMillis = 400
	maxRetryAttemptsCap    = 15
)

// isRetryableGenAIError reports quota pressure and transient server errors.
// 429 RESOURCE_EXHAUSTED is rate/TPM limit or project quota — not a bad request.
func isRetryableGenAIError(err error) bool {
	if err == nil {
		return false
	}
	var ae genai.APIError
	if errors.As(err, &ae) {
		switch ae.Code {
		case http.StatusTooManyRequests, http.StatusInternalServerError,
			http.StatusBadGateway, http.StatusServiceUnavailable, http.StatusGatewayTimeout:
			return true
		}
		if strings.EqualFold(ae.Status, "RESOURCE_EXHAUSTED") ||
			strings.EqualFold(ae.Status, "UNAVAILABLE") {
			return true
		}
	}
	// Wrapped or non-typed errors from SDK / HTTP stack
	s := err.Error()
	return strings.Contains(s, "429") ||
		strings.Contains(s, "RESOURCE_EXHAUSTED") ||
		strings.Contains(s, "503") ||
		strings.Contains(s, "UNAVAILABLE")
}

func effectiveRetryCount() int {
	n := GoogleGenAIConfig.RetryMax
	if n < 0 {
		n = 0
	}
	if n > maxRetryAttemptsCap {
		n = maxRetryAttemptsCap
	}
	return n
}

// total tries = 1 + effectiveRetryCount()
func retryBaseDuration() time.Duration {
	ms := GoogleGenAIConfig.RetryBaseMillis
	if ms < 50 {
		ms = 50
	}
	return time.Duration(ms) * time.Millisecond
}

func retryDelay(attemptZeroBased int) time.Duration {
	base := retryBaseDuration()
	shift := attemptZeroBased
	if shift > 6 {
		shift = 6
	}
	d := base * time.Duration(uint64(1)<<uint(shift))
	const max = 30 * time.Second
	if d > max {
		d = max
	}
	return d + time.Duration(rand.Int64N(350))*time.Millisecond
}

func withGenAIRetryResp[T any](ctx context.Context, op string, fn func() (T, error)) (T, error) {
	retries := effectiveRetryCount()
	attempts := 1 + retries
	var zero T
	var lastErr error
	for i := 0; i < attempts; i++ {
		if i > 0 {
			delay := retryDelay(i - 1)
			slog.InfoContext(ctx, "p_google_genai: retrying transient API error",
				"op", op,
				"attempt", i+1,
				"maxAttempts", attempts,
				"delay", delay.String(),
				"error", lastErr.Error())
			select {
			case <-ctx.Done():
				return zero, fmt.Errorf("%w: last genai error: %v", ctx.Err(), lastErr)
			case <-time.After(delay):
			}
		}
		v, err := fn()
		if err == nil {
			return v, nil
		}
		lastErr = err
		if !isRetryableGenAIError(err) {
			return zero, err
		}
	}
	return zero, lastErr
}

func withGenAIRetry(ctx context.Context, op string, fn func() error) error {
	retries := effectiveRetryCount()
	attempts := 1 + retries
	var lastErr error
	for i := 0; i < attempts; i++ {
		if i > 0 {
			delay := retryDelay(i - 1)
			slog.InfoContext(ctx, "p_google_genai: retrying transient API error",
				"op", op,
				"attempt", i+1,
				"maxAttempts", attempts,
				"delay", delay.String(),
				"error", lastErr.Error())
			select {
			case <-ctx.Done():
				return fmt.Errorf("%w: last genai error: %v", ctx.Err(), lastErr)
			case <-time.After(delay):
			}
		}
		err := fn()
		if err == nil {
			return nil
		}
		lastErr = err
		if !isRetryableGenAIError(err) {
			return err
		}
	}
	return lastErr
}
