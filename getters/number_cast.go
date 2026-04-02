package getters

import "context"

func NumberCast[T, V Number](g Getter[V]) Getter[T] {
	var zero T
	return func(ctx context.Context) (T, error) {
		value, err := g(ctx)
		if err != nil {
			return zero, err
		}
		return T(value), nil
	}
}
