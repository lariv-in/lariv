package getters

import "context"

func Match[T comparable, V any](g Getter[T], cases map[T]Getter[V], missingError Getter[error]) Getter[V] {
	var zero V
	return func(ctx context.Context) (V, error) {
		value, err := g(ctx)
		if err != nil {
			return zero, err
		}
		if caseGetter, ok := cases[value]; ok {
			return caseGetter(ctx)
		}
		missingErr, err := missingError(ctx)
		if err != nil {
			return zero, err
		}
		return zero, missingErr
	}
}
