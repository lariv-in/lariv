package getters

import (
	"context"
	"encoding/json"
	"strings"

	"gorm.io/datatypes"
)

// JSONArray parses JSON from a string getter and returns datatypes.JSON. The top-level value must be a JSON array.
// Empty or whitespace-only input yields "[]".
func JSONArray[T any](g Getter[string]) Getter[datatypes.JSON] {
	return func(ctx context.Context) (datatypes.JSON, error) {
		s, err := g(ctx)
		if err != nil {
			return nil, err
		}
		s = strings.TrimSpace(s)
		if s == "" {
			return datatypes.JSON("[]"), nil
		}
		var a []T
		if err := json.Unmarshal([]byte(s), &a); err != nil {
			return nil, err
		}
		b, err := json.Marshal(a)
		if err != nil {
			return nil, err
		}
		return datatypes.JSON(b), nil
	}
}
