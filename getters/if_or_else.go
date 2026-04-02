package getters

import "context"

// IfOrElse returns a Getter that invokes g when g is non-nil and returns a non-zero value without error;
// otherwise it invokes elseGetter. If elseGetter is nil in those fallback cases, returns the zero value of T.
func IfOrElse[T comparable](g, elseGetter Getter[T]) Getter[T] {
	var zero T
	return func(ctx context.Context) (T, error) {
		if g != nil {
			value, err := g(ctx)
			if err == nil && value != zero {
				return value, nil
			}
		}
		if elseGetter != nil {
			return elseGetter(ctx)
		}
		return zero, nil
	}
}
