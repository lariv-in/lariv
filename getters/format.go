package getters

import (
	"context"
	"fmt"
)

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
