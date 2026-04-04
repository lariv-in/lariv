package p_nirmancampus_assignmentsubmissions

import (
	"fmt"
	"log/slog"
	"net/http"

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

	rawUser := ctx.Value("$user")
	if rawUser == nil {
		slog.Error("AssignmentSubmissionScopeByRole: missing $user in context – auth layer not applied?")
		panic("AssignmentSubmissionScopeByRole: $user is nil in context")
	}
	user, ok := rawUser.(p_users.User)
	if !ok {
		slog.Error("AssignmentSubmissionScopeByRole: $user has unexpected type", "type", fmt.Sprintf("%T", rawUser))
		panic("AssignmentSubmissionScopeByRole: $user has wrong type in context")
	}

	rawRole := ctx.Value("$role")
	if rawRole == nil {
		slog.Error("AssignmentSubmissionScopeByRole: missing $role in context – auth layer not applied?")
		panic("AssignmentSubmissionScopeByRole: $role is nil in context")
	}
	roleName, ok := rawRole.(string)
	if !ok {
		slog.Error("AssignmentSubmissionScopeByRole: $role has unexpected type", "type", fmt.Sprintf("%T", rawRole))
		panic("AssignmentSubmissionScopeByRole: $role has wrong type in context")
	}

	dbVal := ctx.Value("$db")
	db, ok := dbVal.(*gorm.DB)
	if !ok || db == nil {
		slog.Error("AssignmentSubmissionScopeByRole: missing or invalid $db in context", "type", fmt.Sprintf("%T", dbVal))
		panic("AssignmentSubmissionScopeByRole: $db is nil or wrong type in context")
	}

	switch roleName {
	case "superuser", "admin":
		return query
	case "student":
		studentSub := db.Model(&p_nirmancampus_students.Student{}).
			Select("id").
			Where("user_id = ?", user.ID)
		academicRecordSub := db.Table("academic_records").
			Select("id").
			Where("student_id IN (?)", studentSub)
		return query.Where("academic_record_id IN (?)", academicRecordSub)
	default:
		return query.Where("1 = 0")
	}
}

var AssignmentSubmissionScopeByRole views.QueryPatcher[AssignmentSubmission] = assignmentSubmissionScopeByRole{}
