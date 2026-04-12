package getters

import (
	"context"
	"encoding/json"
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

		detail, err := json.Marshal(map[string]any{
			"name":    name,
			"value":   value,
			"display": display,
		})
		if err != nil {
			return "", err
		}
		// JSON object is a valid JS object literal here; gomponents.Attr HTML-escapes the full @click value once.
		// Close the dialog that contains this row (not "last body dialog") so nested/stacked modals cannot remove the wrong one.
		js := fmt.Sprintf("$dispatch('fk-select', %s); $event.currentTarget.closest('dialog.modal')?.remove()", detail)
		return js, nil
	}
}
