package getters

import (
	"context"
	"errors"
)

// Invokes the getter, if it is not nil and returns a non-nil value and does not error out, calls the builder. Otherwise returns the zero value of T.
func If[T any, V comparable](g Getter[V], ctx context.Context, builder func(context.Context, V) (T, error)) (T, error) {
	var zero T
	var zeroV V
	if g == nil {
		return zero, errors.New("Getter is nil")
	}
	value, err := g(ctx)
	if err != nil {
		return zero, err
	}
	if value == zeroV {
		return zero, errors.New("Value is nil")
	}
	return builder(ctx, value)
}
