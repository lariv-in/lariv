package p_nirmancampus_programs

import (
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/lariv-in/lago/plugins/p_nirmancampus_students"
	"github.com/lariv-in/lago/plugins/p_users"
	"github.com/lariv-in/lago/views"
	"gorm.io/gorm"
)

func fieldDBName[T any](db *gorm.DB, fieldName string) (string, bool) {
	stmt := &gorm.Statement{DB: db}
	if err := stmt.Parse(new(T)); err != nil {
		return "", false
	}
	if stmt.Schema == nil {
		return "", false
	}
	field := stmt.Schema.LookUpField(fieldName)
	if field == nil {
		return "", false
	}
	return field.DBName, true
}

func queryPatcherUniversity(param string) views.QueryPatcher {
	return func(_ *views.View, r *http.Request, query *gorm.DB) *gorm.DB {
		getMap, ok := r.Context().Value("$get").(map[string]any)
		if !ok {
			return query
		}

		raw, ok := getMap[param]
		if !ok {
			return query
		}
		value, ok := raw.(string)
		if !ok {
			return query
		}
		value = strings.TrimSpace(value)
		if value == "" {
			return query
		}

		col, ok := fieldDBName[Program](query, "University")
		if !ok {
			return query
		}

		return query.Where(col+" = ?", value)
	}
}

// ProgramScopeByRole restricts program queries:
//   - superuser, admin: full queryset
//   - student: programs referenced by any academic record for this user's student row
//   - any other role: empty queryset
//
// Uses table name academic_records to avoid importing the academicrecords plugin (it imports programs).
func ProgramScopeByRole(_ *views.View, r *http.Request, query *gorm.DB) *gorm.DB {
	ctx := r.Context()

	rawUser := ctx.Value("$user")
	if rawUser == nil {
		slog.Error("ProgramScopeByRole: missing $user in context – auth middleware not applied?")
		panic("ProgramScopeByRole: $user is nil in context")
	}
	user, ok := rawUser.(p_users.User)
	if !ok {
		slog.Error("ProgramScopeByRole: $user has unexpected type",
			"type", fmt.Sprintf("%T", rawUser),
		)
		panic("ProgramScopeByRole: $user has wrong type in context")
	}

	rawRole := ctx.Value("$role")
	if rawRole == nil {
		slog.Error("ProgramScopeByRole: missing $role in context – auth middleware not applied?")
		panic("ProgramScopeByRole: $role is nil in context")
	}
	roleName, ok := rawRole.(string)
	if !ok {
		slog.Error("ProgramScopeByRole: $role has unexpected type",
			"type", fmt.Sprintf("%T", rawRole),
		)
		panic("ProgramScopeByRole: $role has wrong type in context")
	}

	dbVal := ctx.Value("$db")
	db, ok := dbVal.(*gorm.DB)
	if !ok || db == nil {
		slog.Error("ProgramScopeByRole: missing or invalid $db in context",
			"type", fmt.Sprintf("%T", dbVal),
		)
		panic("ProgramScopeByRole: $db is nil or wrong type in context")
	}

	switch roleName {
	case "superuser", "admin":
		return query
	case "student":
		studentSub := db.Model(&p_nirmancampus_students.Student{}).
			Select("id").
			Where("user_id = ?", user.ID)
		programSub := db.Table("academic_records").
			Select("program_id").
			Where("student_id IN (?)", studentSub).
			Where("deleted_at IS NULL")
		return query.Where("programs.id IN (?)", programSub)
	default:
		return query.Where("1 = 0")
	}
}
