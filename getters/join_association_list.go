package getters

import (
	"context"
	"fmt"
	"reflect"
	"sync"

	"gorm.io/gorm"
)

var (
	schemaFieldCache   = make(map[reflect.Type]map[string]string)
	schemaFieldCacheMu sync.RWMutex
)

// JoinAssociationList fetches related records through a join model.
// ownerField and targetField are struct field names on the join model (for example,
// "CourseID" and "TeacherID"). When order is empty, join-row order is preserved.
func JoinAssociationList[TJoin, TTarget any](ownerIDGetter Getter[uint], ownerField, targetField, order string, preloads ...string) Getter[[]TTarget] {
	return func(ctx context.Context) ([]TTarget, error) {
		ownerID, err := ownerIDGetter(ctx)
		if err != nil {
			return nil, err
		}
		if ownerID == 0 {
			return nil, nil
		}

		db, err := DBFromContext(ctx)
		if err != nil {
			return nil, err
		}

		ownerDBName, err := schemaFieldDBName[TJoin](db, ownerField)
		if err != nil {
			return nil, err
		}

		chain := gorm.G[TJoin](db).Where(ownerDBName+" = ?", ownerID)
		if order == "" {
			chain = chain.Order(ownerDBName + " ASC")
		} else {
			chain = chain.Order(order)
		}
		joinRows, err := chain.Find(ctx)
		if err != nil {
			return nil, err
		}
		targetIDs := make([]uint, 0, len(joinRows))
		for i := range joinRows {
			id, ok := uintFromMapField(MapFromStruct(joinRows[i]), targetField)
			if ok {
				targetIDs = append(targetIDs, id)
			}
		}
		if len(targetIDs) == 0 {
			return nil, nil
		}

		return AssociationList[TTarget](Static(targetIDs), order, preloads...)(ctx)
	}
}

func uintFromMapField(m map[string]any, fieldName string) (uint, bool) {
	raw, ok := m[fieldName]
	if !ok || raw == nil {
		return 0, false
	}
	switch v := raw.(type) {
	case uint:
		return v, true
	case uint8:
		return uint(v), true
	case uint16:
		return uint(v), true
	case uint32:
		return uint(v), true
	case uint64:
		return uint(v), true
	case int:
		if v > 0 {
			return uint(v), true
		}
	case int8:
		if v > 0 {
			return uint(v), true
		}
	case int16:
		if v > 0 {
			return uint(v), true
		}
	case int32:
		if v > 0 {
			return uint(v), true
		}
	case int64:
		if v > 0 {
			return uint(v), true
		}
	case *uint:
		if v != nil {
			return *v, true
		}
	case *uint64:
		if v != nil {
			return uint(*v), true
		}
	}
	return 0, false
}

func schemaFieldDBName[T any](db *gorm.DB, fieldName string) (string, error) {
	tType := reflect.TypeOf((*T)(nil)).Elem()
	schemaFieldCacheMu.RLock()
	m, ok := schemaFieldCache[tType]
	var dbName string
	var present bool
	if ok {
		dbName, present = m[fieldName]
	}
	schemaFieldCacheMu.RUnlock()
	if present {
		return dbName, nil
	}

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

	schemaFieldCacheMu.Lock()
	m, ok = schemaFieldCache[tType]
	if !ok {
		m = make(map[string]string)
		schemaFieldCache[tType] = m
	}
	m[fieldName] = field.DBName
	schemaFieldCacheMu.Unlock()

	return field.DBName, nil
}
