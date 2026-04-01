package getters

import "context"

// BoolNot negates a boolean getter. The result is Getter[any] for use with ShowIf and similar.
func BoolNot[T ~bool](g Getter[T]) Getter[any] {
	return func(ctx context.Context) (any, error) {
		v, err := g(ctx)
		if err != nil {
			return nil, err
		}
		return !v, nil
	}
}
