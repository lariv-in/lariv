package getters

import (
	"context"
	"fmt"
)

// Navigate returns an Alpine @click expression that performs HTMX navigation.
// urlFormat and getters work like GetterFormat to produce the URL per-row.
func Navigate(urlFormat string, getters ...Getter[any]) Getter[string] {
	urlGetter := Format(urlFormat, getters...)
	return func(ctx context.Context) (string, error) {
		url, err := IfOr(urlGetter, ctx, "")
		if err != nil {
			return "", err
		}
		// Need to fix this so it uses htmx
		return fmt.Sprintf("htmx.ajax('GET', '%v', {target: 'body', swap: 'outerHTML'})", url), nil
	}
}
