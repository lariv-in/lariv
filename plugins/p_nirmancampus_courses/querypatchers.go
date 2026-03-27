package p_nirmancampus_courses

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/lariv-in/lago/plugins/p_nirmancampus_students"
	"github.com/lariv-in/lago/plugins/p_users"
	"github.com/lariv-in/lago/views"
	"gorm.io/gorm"
)

// CourseScopeByRole restricts course queries:
//   - superuser, admin: full queryset
//   - student: courses linked to any of this user's academic records (via academic_record_courses)
//   - any other role: empty queryset
//
// Table/column names match GORM defaults for AcademicRecord many2many on courses; the academicrecords
// plugin is not imported here to avoid a module import cycle (academicrecords → courses).
func CourseScopeByRole(_ *views.View, r *http.Request, query *gorm.DB) *gorm.DB {
	ctx := r.Context()

	rawUser := ctx.Value("$user")
	if rawUser == nil {
		slog.Error("CourseScopeByRole: missing $user in context – auth middleware not applied?")
		panic("CourseScopeByRole: $user is nil in context")
	}
	user, ok := rawUser.(p_users.User)
	if !ok {
		slog.Error("CourseScopeByRole: $user has unexpected type",
			"type", fmt.Sprintf("%T", rawUser),
		)
		panic("CourseScopeByRole: $user has wrong type in context")
	}

	rawRole := ctx.Value("$role")
	if rawRole == nil {
		slog.Error("CourseScopeByRole: missing $role in context – auth middleware not applied?")
		panic("CourseScopeByRole: $role is nil in context")
	}
	roleName, ok := rawRole.(string)
	if !ok {
		slog.Error("CourseScopeByRole: $role has unexpected type",
			"type", fmt.Sprintf("%T", rawRole),
		)
		panic("CourseScopeByRole: $role has wrong type in context")
	}

	dbVal := ctx.Value("$db")
	db, ok := dbVal.(*gorm.DB)
	if !ok || db == nil {
		slog.Error("CourseScopeByRole: missing or invalid $db in context",
			"type", fmt.Sprintf("%T", dbVal),
		)
		panic("CourseScopeByRole: $db is nil or wrong type in context")
	}

	switch roleName {
	case "superuser", "admin":
		return query
	case "student":
		studentSub := db.Model(&p_nirmancampus_students.Student{}).
			Select("id").
			Where("user_id = ?", user.ID)
		courseSub := db.Table("academic_record_courses").
			Select("academic_record_courses.course_id").
			Joins("JOIN academic_records ON academic_records.id = academic_record_courses.academic_record_id AND academic_records.deleted_at IS NULL").
			Where("academic_records.student_id IN (?)", studentSub)
		return query.Where("courses.id IN (?)", courseSub)
	default:
		return query.Where("1 = 0")
	}
}
