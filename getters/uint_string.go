package getters

import (
	"context"
	"strconv"
)

// UintString converts a Getter[uint] to Getter[string] by formatting the uint.
// Errors from the underlying getter are propagated.
func UintString(g Getter[uint]) Getter[string] {
	return func(ctx context.Context) (string, error) {
		v, err := g(ctx)
		if err != nil {
			return "", err
		}
		return strconv.FormatUint(uint64(v), 10), nil
	}
}
