package getters

import "context"

// Any converts a getter with a concrete type T to a Getter[any].
// This is useful for chaining or when type erasure is needed to satisfy interfaces.
func Any[T any](g Getter[T]) Getter[any] {
	return func(ctx context.Context) (any, error) {
		return g(ctx)
	}
}
