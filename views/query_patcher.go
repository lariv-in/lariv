package views

import (
	"log/slog"
	"net/http"

	"github.com/lariv-in/lago/components"
	"gorm.io/gorm"
)

type QueryPatcher = func(view *View, r *http.Request, db *gorm.DB) *gorm.DB

func QueryPatcherPreload(field string) QueryPatcher {
	return func(v *View, r *http.Request, query *gorm.DB) *gorm.DB {
		return query.Preload(field)
	}
}

func QueryPatcherOrderBy(order string) QueryPatcher {
	return func(v *View, r *http.Request, query *gorm.DB) *gorm.DB {
		return query.Order(order)
	}
}

func joinFilterFieldDBName[T any](db *gorm.DB, fieldName string) (string, bool) {
	stmt := &gorm.Statement{DB: db}
	if err := stmt.Parse(new(T)); err != nil {
		slog.Error("QueryPatcherJoinFilter schema parse failed", "field", fieldName, "error", err)
		return "", false
	}
	if stmt.Schema == nil {
		slog.Error("QueryPatcherJoinFilter schema missing", "field", fieldName)
		return "", false
	}
	field := stmt.Schema.LookUpField(fieldName)
	if field == nil {
		slog.Error("QueryPatcherJoinFilter field missing", "field", fieldName)
		return "", false
	}
	return field.DBName, true
}

func joinFilterIDs(raw any) []uint {
	switch typed := raw.(type) {
	case components.AssociationIDs:
		return typed.IDs
	case *components.AssociationIDs:
		if typed == nil {
			return nil
		}
		return typed.IDs
	case []uint:
		return typed
	default:
		return nil
	}
}

// QueryPatcherJoinFilter filters the current model by a related ID set through a join model.
// param is the filter form field name stored under $get. ownerField and relatedField are
// join-model struct field names, for example "CourseID" and "TeacherID".
func QueryPatcherJoinFilter[TJoin any](param, ownerField, relatedField string) QueryPatcher {
	return func(v *View, r *http.Request, query *gorm.DB) *gorm.DB {
		getMap, ok := r.Context().Value("$get").(map[string]any)
		if !ok {
			return query
		}
		raw, ok := getMap[param]
		if !ok {
			return query
		}
		ids := joinFilterIDs(raw)
		if len(ids) == 0 {
			return query
		}

		ownerDBName, ok := joinFilterFieldDBName[TJoin](query, ownerField)
		if !ok {
			return query
		}
		relatedDBName, ok := joinFilterFieldDBName[TJoin](query, relatedField)
		if !ok {
			return query
		}

		subquery := query.Session(&gorm.Session{NewDB: true}).
			Model(new(TJoin)).
			Select(ownerDBName).
			Where(relatedDBName+" IN ?", ids)

		return query.Where("id IN (?)", subquery)
	}
}
