package p_nirmancampus_assignmentsubmissions

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
