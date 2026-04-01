package getters

import (
	"context"
	"fmt"
	"net/url"
)

func QueryEscape[T comparable](g Getter[T]) Getter[string] {
	var zero T
	return func(ctx context.Context) (string, error) {
		value, err := IfOr(g, ctx, zero)
		if err != nil {
			return "", err
		}
		return url.QueryEscape(fmt.Sprintf("%v", value)), nil
	}
}
