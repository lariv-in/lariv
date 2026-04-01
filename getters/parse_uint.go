package getters

import (
	"context"
	"strconv"
)

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
