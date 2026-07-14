package getters

import (
	"context"
	"strconv"
)

// ParseInt parses a string resolved by the underlying getter into an int.
// It returns an error if the underlying getter fails or if the string cannot
// be parsed as a base-10 integer. This is often used in combination with [NumberCast].
func ParseInt(g Getter[string]) Getter[int] {
	return func(ctx context.Context) (int, error) {
		s, err := g(ctx)
		if err != nil {
			return 0, err
		}
		i, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return 0, err
		}
		return int(i), nil
	}
}
