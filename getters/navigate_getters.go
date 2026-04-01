package getters

import (
	"context"
	"fmt"
)

// NavigateGetter is like GetterNavigate but takes a pre-built Getter for the URL.
func NavigateGetter[T comparable](urlGetter Getter[T]) Getter[string] {
	var zero T
	return func(ctx context.Context) (string, error) {
		url, err := IfOr(urlGetter, ctx, zero)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("htmx.ajax('GET', '%v', {target: 'body', swap: 'outerHTML'})", url), nil
	}
}
