package getters

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"

	"gorm.io/gorm"
)

// ForeignKey fetches a related model T by its primary key and returns a specific field.
// foreignKeyGetter resolves the FK value (e.g. GetterKey("$in.RoleID")).
// fieldPath is the dot-separated path into the related model's map (e.g. "Name").
func ForeignKey[T any, K comparable, V any](foreignKeyGetter Getter[K], fieldPath string) Getter[V] {
	var zeroK K
	var zeroV V
	return func(ctx context.Context) (V, error) {
		fkValue, err := IfOr(foreignKeyGetter, ctx, zeroK)
		if err != nil {
			return zeroV, err
		}

		db, err := DBFromContext(ctx)
		if err != nil {
			return zeroV, err
		}

		instance, err := gorm.G[T](db).Where("id = ?", fkValue).First(ctx)
		if err != nil {
			return zeroV, err
		}

		// Convert to map and walk the field path
		m := MapFromStruct(&instance)
		parts := strings.Split(fieldPath, ".")
		var value any = m
		for _, part := range parts {
			mp, ok := value.(map[string]any)
			if !ok {
				return zeroV, errors.New("Couldn't convert the related field struct to map")
			}
			value, ok = mp[part]
			if !ok {
				return zeroV, errors.New("Couldn't find the key in the struct")
			}
		}
		v, ok := value.(V)
		if !ok {
			return zeroV, fmt.Errorf("Value for key %s found, but the type of value in context was %v, expected %v", fieldPath, reflect.TypeOf(value), reflect.TypeOf(zeroV))
		}
		return v, nil
	}
}
