package getters

import "context"

// Map applies a mapping function to the value resolved by the underlying getter,
// returning a new Getter. It behaves similarly to map operations in functional languages.
// If either the underlying getter or the mapping function returns an error, the error
// is propagated immediately.
//
// Example:
//
//	intGetter := getters.Static(42)
//	stringGetter := getters.Map(intGetter, func(ctx context.Context, v int) (string, error) {
//		return fmt.Sprintf("value: %d", v), nil
//	})
func Map[T, V any](g Getter[T], f func(context.Context, T) (V, error)) Getter[V] {
	var zero V
	return func(ctx context.Context) (V, error) {
		value, err := g(ctx)
		if err != nil {
			return zero, err
		}
		return f(ctx, value)
	}
}
