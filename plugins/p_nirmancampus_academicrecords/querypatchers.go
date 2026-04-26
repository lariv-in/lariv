package p_nirmancampus_academicrecords

import (
	"context"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/lariv-in/lago/getters"
	sessions "github.com/lariv-in/lago/plugins/p_nirmancampus_sessions"
	"github.com/lariv-in/lago/plugins/p_nirmancampus_students"
	"github.com/lariv-in/lago/plugins/p_users"
	"github.com/lariv-in/lago/registry"
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

// selectedAcademicRecordSessionFilter returns whether to restrict the query to a single
// session id, and that id when restrict is true. When restrict is false, the list shows
// all sessions (user chose "—" in the environment selector).
func selectedAcademicRecordSessionFilter(db *gorm.DB, ctx context.Context) (id uint, restrict bool) {
	envMap, ok := ctx.Value("$environment").(map[string]string)
	if !ok {
		id, err := sessions.DefaultAdmissionSessionID(db)
		if err != nil {
			slog.Error("selectedAcademicRecordSessionFilter: no default session",
				"error", err,
			)
			return 0, false
		}
		return id, true
	}
	raw, inMap := envMap[academicRecordsEnvironmentSessionKey]
	if !inMap {
		id, err := sessions.DefaultAdmissionSessionID(db)
		if err != nil {
			slog.Error("selectedAcademicRecordSessionFilter: no default session",
				"error", err,
			)
			return 0, false
		}
		return id, true
	}
	if strings.TrimSpace(raw) == "" {
		return 0, false
	}
	parsed, err := strconv.ParseUint(strings.TrimSpace(raw), 10, 64)
	if err != nil || parsed == 0 {
		slog.Error("selectedAcademicRecordSessionFilter: invalid session id in environment",
			"raw", raw,
			"error", err,
		)
		id, defErr := sessions.DefaultAdmissionSessionID(db)
		if defErr != nil {
			return 0, false
		}
		return id, true
	}
	return uint(parsed), true
}

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

// AcademicRecordQueryPatchersAssignmentSubmissionInput loads one [AcademicRecord] for
// [components.InputForeignKey] display in other plugins (preloads + same role scope as list/detail).
var AcademicRecordQueryPatchersAssignmentSubmissionInput = views.QueryPatchers[AcademicRecord]{
	registry.Pair[string, views.QueryPatcher[AcademicRecord]]{Key: "academicrecords.preload", Value: views.QueryPatcherPreload[AcademicRecord]{Fields: []string{"Student", "Program", "AdmissionSession"}}},
	registry.Pair[string, views.QueryPatcher[AcademicRecord]]{Key: "academicrecords.scope_by_role", Value: AcademicRecordScopeByRole},
}

// AcademicRecordQueryPatchersBulkModal preloads student + course pools for the assignment-submissions bulk-create flow.
var AcademicRecordQueryPatchersBulkModal = views.QueryPatchers[AcademicRecord]{
	registry.Pair[string, views.QueryPatcher[AcademicRecord]]{Key: "academicrecords.preload", Value: views.QueryPatcherPreload[AcademicRecord]{Fields: []string{"Student", "CompulsoryCourses", "OptionalCourses"}}},
	registry.Pair[string, views.QueryPatcher[AcademicRecord]]{Key: "academicrecords.scope_by_role", Value: AcademicRecordScopeByRole},
}
