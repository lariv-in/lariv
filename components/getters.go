package components

import (
	"context"
	"fmt"
	"strings"
)

type Getter func(context.Context) any

func GetterStatic(value any) Getter {
	return func(ctx context.Context) any {
		return value
	}
}

func GetterKey(key string) Getter {
	return func(ctx context.Context) any {
		parts := strings.Split(key, ".")
		value := ctx.Value(parts[0])
		for i := 1; i < len(parts); i++ {
			if value == nil {
				return nil
			}
			m, ok := value.(map[string]any)
			if !ok {
				return nil
			}
			value = m[parts[i]]
		}
		return value
	}
}

func GetterNil() Getter {
	return func(ctx context.Context) any {
		return nil
	}
}

func GetterFormat(format string, g ...Getter) Getter {
	return func(ctx context.Context) any {
		values := []any{}
		for _, getter := range g {
			values = append(values, IfOrGetter(getter, ctx, ""))
		}
		return fmt.Sprintf(format, values...)
	}
}

// Invokes the getter, if it is not nil and returns a non-nil value, returns that value. Otherwise returns the defaultValue.
func IfOrGetter(g Getter, ctx context.Context, defaultValue any) any {
	if g == nil {
		return defaultValue
	}
	value := g(ctx)
	if value == nil {
		return defaultValue
	}
	return value
}

// Invokes the getter, if it is not nil and returns a non-nil value, calls the builder. Otherwise returns the zero value of T.
func GetterIf[T any](g Getter, ctx context.Context, builder func(context.Context, any) T) T {
	var zero T
	if g == nil {
		return zero
	}
	value := g(ctx)
	if value == nil {
		return zero
	}
	return builder(ctx, value)
}
