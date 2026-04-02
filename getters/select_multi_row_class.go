package getters

import (
	"context"
	"fmt"
)

// SelectMultiRowClass returns an Alpine :class expression that derives a row's
// selected styling from the enclosing InputManyToMany items array.
func SelectMultiRowClass[T comparable](valueGetter Getter[T]) Getter[string] {
	var zeroT T
	return func(ctx context.Context) (string, error) {
		value, err := IfOr(valueGetter, ctx, zeroT)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf(
			`items.some(item => item.Key === %q) ? 'bg-success text-success-content hover:bg-success border-success' : 'hover:bg-base-200'`,
			fmt.Sprint(value),
		), nil
	}
}
