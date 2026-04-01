package getters

import "context"

func Nil[T any]() Getter[T] {
	var zero T
	return func(ctx context.Context) (T, error) {
		return zero, nil
	}
}
