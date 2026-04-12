package getters

import (
	"context"
	"fmt"
	"strconv"
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

		vStr := fmt.Sprint(value)
		dStr := fmt.Sprint(display)

		// Double-quoted JS strings (strconv.Quote); gomponents.Attr HTML-escapes the full @click value once.
		// Close the dialog that contains this row (not "last body dialog") so nested/stacked modals cannot remove the wrong one.
		js := fmt.Sprintf("$dispatch('fk-select', {name:%s,value:%s,display:%s}); $event.currentTarget.closest('dialog.modal')?.remove()",
			strconv.Quote(name),
			strconv.Quote(vStr),
			strconv.Quote(dStr),
		)
		return js, nil
	}
}
