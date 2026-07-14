package getters

import (
	"context"
	"fmt"
)

// Format formats a string analogous to [fmt.Sprintf]. Instead of accepting variadic
// any values, it accepts variadic [Getter]s of any, resolving them using the provided context.
// If any of the getters return an error during resolution, that error is returned immediately.
func Format(format string, g ...Getter[any]) Getter[string] {
	return func(ctx context.Context) (string, error) {
		values := []any{}
		for _, getter := range g {
			v, err := IfOr(getter, ctx, "")
			if err != nil {
				return "", err
			}
			values = append(values, v)
		}
		return fmt.Sprintf(format, values...), nil
	}
}
