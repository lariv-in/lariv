package getters

import (
	"context"
	"fmt"
)

// SelectMultiRowClass returns an Alpine :class expression that derives a row's
// selected styling from the InputManyToMany store for the named field.
func SelectMultiRowClass[T comparable](nameGetter Getter[string], valueGetter Getter[T]) Getter[string] {
	var zeroT T
	return func(ctx context.Context) (string, error) {
		name, err := IfOr(nameGetter, ctx, "")
		if err != nil {
			return "", err
		}
		value, err := IfOr(valueGetter, ctx, zeroT)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf(
			`((Alpine.store('m2mSelections') && Alpine.store('m2mSelections')[%q]) || []).some(item => item.Key === %q) ? 'bg-success text-success-content hover:bg-success border-success' : 'hover:bg-base-200'`,
			name,
			fmt.Sprint(value),
		), nil
	}
}
