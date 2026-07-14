package getters

import (
	"context"
	"strconv"
)

// ParseUint parses a string resolved by the underlying getter into a uint.
// It returns an error if the underlying getter fails or if the string cannot
// be parsed as a base-10 unsigned integer. This is often used in combination with [NumberCast].
func ParseUint(g Getter[string]) Getter[uint] {
	return func(ctx context.Context) (uint, error) {
		s, err := g(ctx)
		if err != nil {
			return 0, err
		}
		u, err := strconv.ParseUint(s, 10, 64)
		if err != nil {
			return 0, err
		}
		return uint(u), nil
	}
}
