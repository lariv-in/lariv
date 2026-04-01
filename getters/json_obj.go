package getters

import (
	"context"
	"encoding/json"
	"strings"

	"gorm.io/datatypes"
)

// JSONObj parses JSON from a string getter and returns datatypes.JSON. The top-level value must be a JSON object.
// Empty or whitespace-only input yields "{}".
func JSONObj[T any](g Getter[string]) Getter[datatypes.JSON] {
	return func(ctx context.Context) (datatypes.JSON, error) {
		s, err := g(ctx)
		if err != nil {
			return nil, err
		}
		s = strings.TrimSpace(s)
		if s == "" {
			return datatypes.JSON("{}"), nil
		}
		var m T
		if err := json.Unmarshal([]byte(s), &m); err != nil {
			return nil, err
		}
		b, err := json.Marshal(m)
		if err != nil {
			return nil, err
		}
		return datatypes.JSON(b), nil
	}
}
