package getters

import "context"

// Match evaluates a key getter and resolves to the matching case from the cases map.
// It is analogous to a Rust match statement.
// If the key is not found in the cases map, the missingError getter is evaluated
// and returned as the resolution error.
//
// Example:
//
//	statusGetter := getters.Static("admin")
//	viewGetter := getters.Match(
//		statusGetter,
//		map[string]getters.Getter[string]{
//			"admin": getters.Static("Show Admin Dashboard"),
//			"user":  getters.Static("Show User Settings"),
//		},
//		getters.Static(errors.New("unknown role")),
//	)
func Match[T comparable, V any](g Getter[T], cases map[T]Getter[V], missingError Getter[error]) Getter[V] {
	var zero V
	return func(ctx context.Context) (V, error) {
		value, err := g(ctx)
		if err != nil {
			return zero, err
		}
		if caseGetter, ok := cases[value]; ok {
			return caseGetter(ctx)
		}
		missingErr, err := missingError(ctx)
		if err != nil {
			return zero, err
		}
		return zero, missingErr
	}
}
