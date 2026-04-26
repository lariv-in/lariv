package p_nirmancampus_assignmentsubmissions

import (
	"context"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/plugins/p_nirmancampus_academicrecords"
	sessions "github.com/lariv-in/lago/plugins/p_nirmancampus_sessions"
	"github.com/lariv-in/lago/plugins/p_nirmancampus_students"
	"github.com/lariv-in/lago/plugins/p_users"
	"github.com/lariv-in/lago/views"
	"gorm.io/gorm"
)

type assignmentSubmissionScopeByRole struct{}

// AssignmentSubmissionScopeByRole restricts queries:
// - superuser/admin: full queryset
// - student: submissions tied to this user's academic records
// - any other role: empty queryset
func (assignmentSubmissionScopeByRole) Patch(_ views.View, r *http.Request, query gorm.ChainInterface[AssignmentSubmission]) gorm.ChainInterface[AssignmentSubmission] {
	ctx := r.Context()
	user, roleName := p_users.UserAndRoleFromContext(ctx, "AssignmentSubmissionScopeByRole")

	db, err := getters.DBFromContext(ctx)
	if err != nil {
		slog.Error("AssignmentSubmissionScopeByRole: db from context", "error", err)
		panic("AssignmentSubmissionScopeByRole: " + err.Error())
	}

	switch roleName {
	case "superuser", "admin":
		return query
	case "student":
		email := strings.TrimSpace(user.Email)
		if email == "" {
			return query.Where("1 = 0")
		}
		studentSub := db.Model(&p_nirmancampus_students.Student{}).
			Select("id").
			Where("email = ?", email)
		academicRecordSub := db.Table("academic_records").
			Select("id").
			Where("student_id IN (?)", studentSub)
		return query.Where("academic_record_id IN (?)", academicRecordSub)
	default:
		return query.Where("1 = 0")
	}
}

var AssignmentSubmissionScopeByRole views.QueryPatcher[AssignmentSubmission] = assignmentSubmissionScopeByRole{}

type assignmentSubmissionListSessionFilter struct{}

func selectedAssignmentSubmissionsSessionFilter(db *gorm.DB, ctx context.Context) (id uint, restrict bool) {
	envMap, ok := ctx.Value("$environment").(map[string]string)
	if !ok {
		id, err := sessions.DefaultAdmissionSessionID(db)
		if err != nil {
			slog.Error("selectedAssignmentSubmissionsSessionFilter: no default session", "error", err)
			return 0, false
		}
		return id, true
	}
	raw, inMap := envMap[assignmentSubmissionsEnvironmentSessionKey]
	if !inMap {
		id, err := sessions.DefaultAdmissionSessionID(db)
		if err != nil {
			slog.Error("selectedAssignmentSubmissionsSessionFilter: no default session", "error", err)
			return 0, false
		}
		return id, true
	}
	if strings.TrimSpace(raw) == "" {
		return 0, false
	}
	parsed, err := strconv.ParseUint(strings.TrimSpace(raw), 10, 64)
	if err != nil || parsed == 0 {
		slog.Error("selectedAssignmentSubmissionsSessionFilter: invalid session id in environment", "raw", raw, "error", err)
		id, defErr := sessions.DefaultAdmissionSessionID(db)
		if defErr != nil {
			return 0, false
		}
		return id, true
	}
	return uint(parsed), true
}

func (assignmentSubmissionListSessionFilter) Patch(_ views.View, r *http.Request, query gorm.ChainInterface[AssignmentSubmission]) gorm.ChainInterface[AssignmentSubmission] {
	db, err := getters.DBFromContext(r.Context())
	if err != nil {
		slog.Error("assignmentSubmissionListSessionFilter: db from context", "error", err)
		return query
	}
	id, restrict := selectedAssignmentSubmissionsSessionFilter(db, r.Context())
	if !restrict {
		return query
	}
	academicRecordSub := db.Model(&p_nirmancampus_academicrecords.AcademicRecord{}).
		Select("id").
		Where("session_id = ?", id)
	return query.Where("academic_record_id IN (?)", academicRecordSub)
}

var AssignmentSubmissionListSessionFilter views.QueryPatcher[AssignmentSubmission] = assignmentSubmissionListSessionFilter{}
