package getters

import "context"

// Static returns a Getter which will always return a static value
// Never errors
func Static[T any](value T) Getter[T] {
	return func(ctx context.Context) (T, error) {
		return value, nil
	}
}
