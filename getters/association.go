package getters

import (
	"context"

	"gorm.io/gorm"
)

// Association fetches a single record based on a foreign key dynamically at render time.
func Association[T, V any](foreignKeyGetter Getter[V]) Getter[T] {
	var zero T
	return func(ctx context.Context) (T, error) {
		fkValue, err := foreignKeyGetter(ctx)
		if err != nil {
			return zero, err
		}

		db, err := DBFromContext(ctx)
		if err != nil {
			return zero, err
		}

		result, err := gorm.G[T](db).Where("id = ?", fkValue).Take(ctx)
		if err != nil {
			return zero, err
		}
		return result, nil
	}
}
