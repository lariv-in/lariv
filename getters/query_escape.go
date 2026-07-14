package getters

import (
	"context"
	"fmt"
	"net/url"
)

// QueryEscape converts the value resolved by the underlying getter to a string
// and escapes it using [url.QueryEscape] so that it is safe to be placed inside
// the query section of a URL.
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
