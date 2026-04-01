package getters

import "context"

// Invokes the getter, if it is not nil and returns a non-nil value, and does not error out, returns that value. Otherwise returns the defaultValue.
func IfOr[T comparable](g Getter[T], ctx context.Context, defaultValue T) (T, error) {
	var zero T
	if g == nil {
		return defaultValue, nil
	}
	value, err := g(ctx)
	if err != nil {
		return defaultValue, nil
	}
	if value == zero {
		return defaultValue, nil
	}
	return value, nil
}
