package getters

import (
	"context"
	"strconv"
)

// IntString converts a Getter[int] to Getter[string] by formatting the int.
// Errors from the underlying getter (e.g. type mismatch) are propagated.
func IntString(g Getter[int]) Getter[string] {
	return func(ctx context.Context) (string, error) {
		v, err := g(ctx)
		if err != nil {
			return "", err
		}
		return strconv.Itoa(v), nil
	}
}
