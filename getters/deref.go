package getters

import "context"

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
