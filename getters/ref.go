package getters

import "context"

// Ref wraps a value getter as a pointer getter. It returns a pointer to a copy of the resolved value.
// This is the inverse of [Deref]: [Deref] unwraps *T → T; Ref wraps T → *T.
func Ref[T any](g Getter[T]) Getter[*T] {
	return func(ctx context.Context) (*T, error) {
		value, err := g(ctx)
		if err != nil {
			return nil, err
		}
		return new(value), nil
	}
}
