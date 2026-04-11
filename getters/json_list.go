package getters

import (
	"context"
	"encoding/json"

	"gorm.io/datatypes"
)

// JSONList parses datatypes.JSON from a getter and returns a []T.
// Empty input yields an empty slice.
func JSONList[T any](g Getter[datatypes.JSON]) Getter[[]T] {
	return func(ctx context.Context) ([]T, error) {
		raw, err := g(ctx)
		if err != nil {
			return nil, err
		}
		if len(raw) == 0 {
			return []T{}, nil
		}
		var items []T
		if err := json.Unmarshal(raw, &items); err != nil {
			return nil, err
		}
		return items, nil
	}
}
