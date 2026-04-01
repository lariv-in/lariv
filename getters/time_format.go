package getters

import (
	"context"
	"time"
)

// TimeFormat converts a Getter[time.Time] to Getter[string] by formatting
// the time using the provided layout. Errors from the underlying getter are
// propagated.
func TimeFormat(layout string, g Getter[time.Time]) Getter[string] {
	return func(ctx context.Context) (string, error) {
		t, err := g(ctx)
		if err != nil {
			return "", err
		}
		if t.IsZero() {
			return "", nil
		}
		return t.Format(layout), nil
	}
}
