package getters

import "context"

func Map[T, V any](g Getter[T], f func(context.Context, T) (V, error)) Getter[V] {
	var zero V
	return func(ctx context.Context) (V, error) {
		value, err := g(ctx)
		if err != nil {
			return zero, err
		}
		return f(ctx, value)
	}
}
