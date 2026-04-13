package p_nirmancampus_academicrecords

import (
	"log/slog"
	"net/http"
	"strings"

	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/plugins/p_nirmancampus_students"
	"github.com/lariv-in/lago/plugins/p_users"
	"github.com/lariv-in/lago/views"
	"gorm.io/gorm"
)

type academicRecordScopeByRole struct{}

// AcademicRecordScopeByRole restricts academic record queries:
// - superuser, admin: full queryset
// - student: only academic records for Student rows whose Email matches the logged-in user's email
// - default (any other role): empty queryset
func (academicRecordScopeByRole) Patch(_ views.View, r *http.Request, query gorm.ChainInterface[AcademicRecord]) gorm.ChainInterface[AcademicRecord] {
	ctx := r.Context()
	user, roleName := p_users.UserAndRoleFromContext(ctx, "AcademicRecordScopeByRole")

	db, err := getters.DBFromContext(ctx)
	if err != nil {
		slog.Error("AcademicRecordScopeByRole: db from context", "error", err)
		panic("AcademicRecordScopeByRole: " + err.Error())
	}

	// AuthenticationLayer sets $role to "superuser" for superusers, else Role.name from DB.
	switch roleName {
	case "superuser":
		return query
	case "admin":
		return query
	case "student":
		email := strings.TrimSpace(user.Email)
		if email == "" {
			return query.Where("1 = 0")
		}
		sub := db.Model(&p_nirmancampus_students.Student{}).Select("id").Where("email = ?", email)
		return query.Where("student_id IN (?)", sub)
	default:
		return query.Where("1 = 0")
	}
}

var AcademicRecordScopeByRole views.QueryPatcher[AcademicRecord] = academicRecordScopeByRole{}

type academicRecordListSessionFilter struct{}

func (academicRecordListSessionFilter) Patch(_ views.View, r *http.Request, query gorm.ChainInterface[AcademicRecord]) gorm.ChainInterface[AcademicRecord] {
	db, err := getters.DBFromContext(r.Context())
	if err != nil {
		slog.Error("academicRecordListSessionFilter: db from context", "error", err)
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
