package p_nirmancampus_academicrecords

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/lariv-in/lago/plugins/p_nirmancampus_students"
	"github.com/lariv-in/lago/plugins/p_users"
	"github.com/lariv-in/lago/views"
	"gorm.io/gorm"
)

type academicRecordScopeByRole struct{}

// AcademicRecordScopeByRole restricts academic record queries:
// - superuser, admin: full queryset
// - student: only academic records for the logged-in user's Student row (read via list/detail/select)
// - default (any other role): empty queryset
func (academicRecordScopeByRole) Patch(_ views.View, r *http.Request, query gorm.ChainInterface[AcademicRecord]) gorm.ChainInterface[AcademicRecord] {
	ctx := r.Context()

	rawUser := ctx.Value("$user")
	if rawUser == nil {
		slog.Error("AcademicRecordScopeByRole: missing $user in context – auth layer not applied?")
		panic("AcademicRecordScopeByRole: $user is nil in context")
	}
	user, ok := rawUser.(p_users.User)
	if !ok {
		slog.Error("AcademicRecordScopeByRole: $user has unexpected type",
			"type", fmt.Sprintf("%T", rawUser),
		)
		panic("AcademicRecordScopeByRole: $user has wrong type in context")
	}

	rawRole := ctx.Value("$role")
	if rawRole == nil {
		slog.Error("AcademicRecordScopeByRole: missing $role in context – auth layer not applied?")
		panic("AcademicRecordScopeByRole: $role is nil in context")
	}
	roleName, ok := rawRole.(string)
	if !ok {
		slog.Error("AcademicRecordScopeByRole: $role has unexpected type",
			"type", fmt.Sprintf("%T", rawRole),
		)
		panic("AcademicRecordScopeByRole: $role has wrong type in context")
	}

	dbVal := ctx.Value("$db")
	db, ok := dbVal.(*gorm.DB)
	if !ok || db == nil {
		slog.Error("AcademicRecordScopeByRole: missing or invalid $db in context",
			"type", fmt.Sprintf("%T", dbVal),
		)
		panic("AcademicRecordScopeByRole: $db is nil or wrong type in context")
	}

	// AuthenticationLayer sets $role to "superuser" for superusers, else Role.name from DB.
	switch roleName {
	case "superuser":
		return query
	case "admin":
		return query
	case "student":
		sub := db.Model(&p_nirmancampus_students.Student{}).Select("id").Where("user_id = ?", user.ID)
		return query.Where("student_id IN (?)", sub)
	default:
		return query.Where("1 = 0")
	}
}

var AcademicRecordScopeByRole views.QueryPatcher[AcademicRecord] = academicRecordScopeByRole{}

type academicRecordListSessionFilter struct{}

func (academicRecordListSessionFilter) Patch(_ views.View, r *http.Request, query gorm.ChainInterface[AcademicRecord]) gorm.ChainInterface[AcademicRecord] {
	dbVal := r.Context().Value("$db")
	db, ok := dbVal.(*gorm.DB)
	if !ok || db == nil {
		slog.Error("academicRecordListSessionFilter: missing or invalid $db in context",
			"type", fmt.Sprintf("%T", dbVal),
		)
		return query
	}
	id, restrict := selectedAcademicRecordSessionFilter(db, r.Context())
	if !restrict {
		return query
	}
	return query.Where("session_id = ?", id)
}

// AcademicRecordListSessionFilter scopes list/select queries to the session chosen in
// the environment cookie (or the default active / latest session).
var AcademicRecordListSessionFilter views.QueryPatcher[AcademicRecord] = academicRecordListSessionFilter{}
