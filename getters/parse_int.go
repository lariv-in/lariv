package getters

import (
	"context"
	"strconv"
)

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
