package getters

import (
	"context"
	"errors"

	"gorm.io/gorm"
)

// Association fetches a single record based on a foreign key dynamically at render time.
func Association[T any, V any](foreignKeyGetter Getter[V]) Getter[T] {
	var zero T
	return func(ctx context.Context) (T, error) {
		fkValue, err := foreignKeyGetter(ctx)
		if err != nil {
			return zero, err
		}

		db, ok := ctx.Value("$db").(*gorm.DB)
		if !ok {
			return zero, errors.New("Couldn't load db connection from context")
		}

		var result T
		if err := db.Model(new(T)).Where("id = ?", fkValue).Take(&result).Error; err != nil {
			return zero, err
		}
		return result, nil
	}
}
