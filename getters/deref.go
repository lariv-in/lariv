package getters

import "context"

// Deref converts a Getter[*T] to a Getter[T]. If the pointer resolved by the
// underlying getter is nil, the zero value of T is returned. This is useful for chaining.
func Deref[T any](g Getter[*T]) Getter[T] {
	var zero T
	return func(ctx context.Context) (T, error) {
		value, err := g(ctx)
		if err != nil {
			return zero, err
		}
		if value == nil {
			return zero, nil
		}
		return *value, nil
	}
}
