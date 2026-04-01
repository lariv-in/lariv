package getters

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"
)

// JoinAssociationList fetches related records through a join model.
// ownerField and targetField are struct field names on the join model (for example,
// "CourseID" and "TeacherID"). When order is empty, join-row order is preserved.
func JoinAssociationList[TJoin any, TTarget any](ownerIDGetter Getter[uint], ownerField, targetField, order string, preloads ...string) Getter[[]TTarget] {
	return func(ctx context.Context) ([]TTarget, error) {
		ownerID, err := ownerIDGetter(ctx)
		if err != nil {
			return nil, err
		}
		if ownerID == 0 {
			return nil, nil
		}

		db, ok := ctx.Value("$db").(*gorm.DB)
		if !ok {
			return nil, errors.New("Couldn't load db connection from context")
		}

		ownerDBName, err := schemaFieldDBName[TJoin](db, ownerField)
		if err != nil {
			return nil, err
		}
		targetDBName, err := schemaFieldDBName[TJoin](db, targetField)
		if err != nil {
			return nil, err
		}

		joinQuery := db.Model(new(TJoin)).Where(ownerDBName+" = ?", ownerID)
		if order == "" {
			joinQuery = joinQuery.Order(ownerDBName + " ASC")
		}

		var targetIDs []uint
		if err := joinQuery.Pluck(targetDBName, &targetIDs).Error; err != nil {
			return nil, err
		}
		if len(targetIDs) == 0 {
			return nil, nil
		}

		return AssociationList[TTarget](Static(targetIDs), order, preloads...)(ctx)
	}
}

func schemaFieldDBName[T any](db *gorm.DB, fieldName string) (string, error) {
	stmt := &gorm.Statement{DB: db}
	if err := stmt.Parse(new(T)); err != nil {
		return "", err
	}
	if stmt.Schema == nil {
		return "", fmt.Errorf("schema not found for %T", new(T))
	}
	field := stmt.Schema.LookUpField(fieldName)
	if field == nil {
		return "", fmt.Errorf("field %q not found for %T", fieldName, new(T))
	}
	return field.DBName, nil
}
