package getters

import "context"

// Nil returns a Getter that always resolves to the zero value of type T and never returns an error.
func Nil[T any]() Getter[T] {
	var zero T
	return func(ctx context.Context) (T, error) {
		return zero, nil
	}
}
