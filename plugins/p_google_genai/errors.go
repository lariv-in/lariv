package p_google_genai

import (
	"errors"
	"strings"

	"google.golang.org/genai"
)

// RetryableQuotaError is true when err is a transient Google GenAI / Vertex quota
// or rate limit (HTTP 429 or RESOURCE_EXHAUSTED). Typical use: backoff and retry
// a stream before any response chunks were emitted.
func RetryableQuotaError(err error) bool {
	var apiErr genai.APIError
	if !errors.As(err, &apiErr) {
		return false
	}
	if apiErr.Code == 429 {
		return true
	}
	return strings.EqualFold(apiErr.Status, "RESOURCE_EXHAUSTED")
}

// DefaultStreamMaxAttempts is the default total tries (initial + retries) for
// streaming calls that backoff on [RetryableQuotaError].
const DefaultStreamMaxAttempts = 4
