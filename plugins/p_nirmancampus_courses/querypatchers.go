package p_nirmancampus_courses

import (
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/plugins/p_nirmancampus_students"
	"github.com/lariv-in/lago/plugins/p_users"
	"github.com/lariv-in/lago/views"
	"gorm.io/gorm"
)

type courseScopeByRole struct{}

// CourseScopeByRole restricts course queries:
//   - superuser, admin: full queryset
//   - student: courses linked to any of this user's academic records (compulsory or optional join tables)
//   - any other role: empty queryset
//
// Join table names match GORM tags on AcademicRecord; the academicrecords plugin is not imported here
// to avoid a module import cycle (academicrecords → courses).
func (courseScopeByRole) Patch(_ views.View, r *http.Request, query gorm.ChainInterface[Course]) gorm.ChainInterface[Course] {
	ctx := r.Context()
	user, roleName := p_users.UserAndRoleFromContext(ctx, "CourseScopeByRole")

	db, err := getters.DBFromContext(ctx)
	if err != nil {
		slog.Error("CourseScopeByRole: db from context", "error", err)
		panic("CourseScopeByRole: " + err.Error())
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
		compulsorySub := db.Table("academic_record_compulsory_courses").
			Select("academic_record_compulsory_courses.course_id").
			Joins("JOIN academic_records ON academic_records.id = academic_record_compulsory_courses.academic_record_id AND academic_records.deleted_at IS NULL").
			Where("academic_records.student_id IN (?)", studentSub)
		optionalSub := db.Table("academic_record_optional_courses").
			Select("academic_record_optional_courses.course_id").
			Joins("JOIN academic_records ON academic_records.id = academic_record_optional_courses.academic_record_id AND academic_records.deleted_at IS NULL").
			Where("academic_records.student_id IN (?)", studentSub)
		return query.Where("(courses.id IN (?) OR courses.id IN (?))", compulsorySub, optionalSub)
	default:
		return query.Where("1 = 0")
	}
}

var CourseScopeByRole views.QueryPatcher[Course] = courseScopeByRole{}

// QueryPatcherMultiSelectPoolCourseIDs restricts the multi-select course list when the request includes
// pool_course_ids (comma-separated course IDs). Used by academic record optional-course pickers. If the
// parameter is present with an empty value, the list is empty. If the parameter is absent, no extra filter applies.
type queryPatcherMultiSelectPoolCourseIDs struct{}

func (queryPatcherMultiSelectPoolCourseIDs) Patch(_ views.View, r *http.Request, query gorm.ChainInterface[Course]) gorm.ChainInterface[Course] {
	raw, ok := r.URL.Query()["pool_course_ids"]
	if !ok || len(raw) == 0 {
		return query
	}
	s := strings.TrimSpace(raw[0])
	if s == "" {
		return query.Where("1 = 0")
	}
	parts := strings.Split(s, ",")
	ids := make([]uint, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		n, err := strconv.ParseUint(p, 10, 64)
		if err != nil {
			slog.Error("QueryPatcherMultiSelectPoolCourseIDs: invalid course id segment", "segment", p, "error", err)
			return query.Where("1 = 0")
		}
		ids = append(ids, uint(n))
	}
	if len(ids) == 0 {
		return query.Where("1 = 0")
	}
	return query.Where("courses.id IN ?", ids)
}

var QueryPatcherMultiSelectPoolCourseIDs views.QueryPatcher[Course] = queryPatcherMultiSelectPoolCourseIDs{}
