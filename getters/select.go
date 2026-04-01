package getters

import (
	"context"
	"fmt"
)

// Select returns an Alpine @click expression that dispatches an 'fk-select' event for single selection.
// name is the input field name. valueGetter and displayGetter resolve per-row.
func Select[T, D comparable](name string, valueGetter Getter[T], displayGetter Getter[D]) Getter[string] {
	var zeroT T
	var zeroD D
	return func(ctx context.Context) (string, error) {
		value, err := IfOr(valueGetter, ctx, zeroT)
		if err != nil {
			return "", err
		}
		display, err := IfOr(displayGetter, ctx, zeroD)
		if err != nil {
			return "", err
		}

		return fmt.Sprintf("$dispatch('fk-select',{name:'%s',value:'%v',display:'%v'})", name, value, display), nil
	}
}
