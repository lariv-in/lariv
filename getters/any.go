package getters

import "context"

func Any[T any](g Getter[T]) Getter[any] {
	return func(ctx context.Context) (any, error) {
		return g(ctx)
	}
}
