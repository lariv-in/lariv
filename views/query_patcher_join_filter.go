package views

import (
	"log/slog"
	"net/http"

	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"gorm.io/gorm"
)

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
// Param is the filter form field name stored under $get. OwnerField and RelatedField are
// join-model struct field names, for example "CourseID" and "TeacherID".
type QueryPatcherJoinFilter[T any, TJoin any] struct {
	Param        string
	OwnerField   string
	RelatedField string
}

func (p QueryPatcherJoinFilter[T, TJoin]) Patch(_ View, r *http.Request, db gorm.ChainInterface[T]) gorm.ChainInterface[T] {
	getMap, ok := r.Context().Value("$get").(map[string]any)
	if !ok {
		return db
	}
	raw, ok := getMap[p.Param]
	if !ok {
		return db
	}
	ids := joinFilterIDs(raw)
	if len(ids) == 0 {
		return db
	}

	dbConn, err := getters.DBFromContext(r.Context())
	if err != nil {
		slog.Error("QueryPatcherJoinFilter: db from context", "error", err)
		return db
	}

	ownerDBName, ok := joinFilterFieldDBName[TJoin](dbConn, p.OwnerField)
	if !ok {
		return db
	}
	relatedDBName, ok := joinFilterFieldDBName[TJoin](dbConn, p.RelatedField)
	if !ok {
		return db
	}

	subquery := dbConn.Session(&gorm.Session{NewDB: true}).
		Model(new(TJoin)).
		Select(ownerDBName).
		Where(relatedDBName+" IN ?", ids)

	return db.Where("id IN (?)", subquery)
}
