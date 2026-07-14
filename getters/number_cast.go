package getters

import "context"

// NumberCast converts a Getter[V] to a Getter[T] by casting the resolved numeric value.
// This is useful for converting between different numeric types, such as uint32 and uint64.
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
