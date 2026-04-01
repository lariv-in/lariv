package getters

import (
	"context"
	"fmt"
)

// SelectMulti returns an Alpine @click expression that dispatches an
// 'fk-multi-select' event for multi-selection inputs. nameGetter resolves the
// input field name (e.g. getters.GetterStatic("Field"), or
// getters.IfOrElseGetter(getters.GetterKey[string]("$get.target_input"), getters.GetterStatic("Field"))
// when the name may come from target_input on the request).
func SelectMulti[T, D comparable](nameGetter Getter[string], valueGetter Getter[T], displayGetter Getter[D]) Getter[string] {
	var zeroT T
	var zeroD D
	return func(ctx context.Context) (string, error) {
		name, err := IfOr(nameGetter, ctx, "")
		if err != nil {
			return "", err
		}
		value, err := IfOr(valueGetter, ctx, zeroT)
		if err != nil {
			return "", err
		}
		display, err := IfOr(displayGetter, ctx, zeroD)
		if err != nil {
			return "", err
		}

		return fmt.Sprintf("$dispatch('fk-multi-select',{name:'%s',value:'%v',display:'%v'})", name, value, display), nil
	}
}
