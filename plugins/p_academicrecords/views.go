package p_academicrecords

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/p_semesters"
	"github.com/lariv-in/lago/p_users"
	"github.com/lariv-in/lago/views"
	"gorm.io/gorm"
)

func parseSemesterEnvID(raw string) (uint, bool) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return 0, false
	}
	if i := strings.IndexByte(raw, ':'); i > 0 {
		raw = raw[:i]
	}
	id, err := strconv.ParseUint(raw, 10, 64)
	if err != nil || id == 0 {
		return 0, false
	}
	return uint(id), true
}

// semesterEnvironmentDefault returns the semester id for the semester whose
// [Start, End] interval contains now (inclusive), or (0, false) if none.
func semesterEnvironmentDefault(db *gorm.DB, now time.Time) (uint, bool) {
	var sem p_semesters.Semester
	err := db.Model(&p_semesters.Semester{}).
		Where(`"start" <= ? AND "end" >= ?`, now, now).
		Order(`"start" ASC`).
		First(&sem).Error
	if err != nil {
		return 0, false
	}
	return sem.ID, true
}

// academicRecordsListSemesterEnvQueryPatcher scopes the list view to the semester
// selected in the environment cookie (components.Environment[uint] semester key).
func academicRecordsListSemesterEnvQueryPatcher(_ *views.View, r *http.Request, query *gorm.DB) *gorm.DB {
	envMap, ok := r.Context().Value("$environment").(map[string]string)
	if !ok {
		return query
	}
	raw, ok := envMap["semester"]
	if !ok || strings.TrimSpace(raw) == "" {
		db, dbOK := r.Context().Value("$db").(*gorm.DB)
		if !dbOK || db == nil {
			return query
		}
		id, found := semesterEnvironmentDefault(db, time.Now())
		if !found {
			return query
		}
		raw = fmt.Sprintf("%d", id)
	}
	semesterID, ok := parseSemesterEnvID(raw)
	if !ok {
		return query
	}
	return query.Where("semester_id = ?", semesterID)
}

func init() {
	// List view
	lago.RegistryView.Register("academicrecords.ListView",
		views.ListView[AcademicRecord]("academicrecords")(
			lago.GetPageView("academicrecords.AcademicRecordTable"),
		).
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware).
			WithQueryPatcher("academicrecords.preload_student_user", views.QueryPatcherPreload("Student.User")).
			WithQueryPatcher("academicrecords.preload_semester", views.QueryPatcherPreload("Semester")).
			WithQueryPatcher("academicrecords.scope_by_role", AcademicRecordScopeByRole).
			WithQueryPatcher("academicrecords.filter_env_semester", academicRecordsListSemesterEnvQueryPatcher),
	)

	// Detail view
	lago.RegistryView.Register("academicrecords.DetailView",
		views.DetailView[AcademicRecord]("academicrecord")(
			lago.GetPageView("academicrecords.AcademicRecordDetail"),
		).
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware).
			WithQueryPatcher("academicrecords.preload_student_user", views.QueryPatcherPreload("Student.User")).
			WithQueryPatcher("academicrecords.preload_semester", views.QueryPatcherPreload("Semester")).
			WithQueryPatcher("academicrecords.scope_by_role", AcademicRecordScopeByRole),
	)

	// Create view
	lago.RegistryView.Register("academicrecords.CreateView",
		views.CreateView[AcademicRecord](
			lago.GetterRoutePath("academicrecords.DetailRoute", map[string]getters.Getter[any]{
				"id": getters.GetterAny(getters.GetterKey[uint]("$id")),
			}),
		)(
			lago.GetPageView("academicrecords.AcademicRecordCreateForm"),
		).
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware),
	)

	// Update view
	lago.RegistryView.Register("academicrecords.UpdateView",
		views.DetailView[AcademicRecord]("academicrecord")(
			views.UpdateView[AcademicRecord](
				lago.GetterRoutePath("academicrecords.DetailRoute", map[string]getters.Getter[any]{
					"id": getters.GetterAny(getters.GetterKey[uint]("$id")),
				}),
			)(
				lago.GetPageView("academicrecords.AcademicRecordUpdateForm"),
			),
		).
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware).
			WithQueryPatcher("academicrecords.preload_student_user", views.QueryPatcherPreload("Student.User")).
			WithQueryPatcher("academicrecords.preload_semester", views.QueryPatcherPreload("Semester")).
			WithQueryPatcher("academicrecords.scope_by_role", AcademicRecordScopeByRole),
	)

	// Delete view
	lago.RegistryView.Register("academicrecords.DeleteView",
		views.DetailView[AcademicRecord]("academicrecord")(
			views.DeleteView[AcademicRecord](
				lago.GetterRoutePath("academicrecords.DefaultRoute", nil),
			)(
				lago.GetPageView("academicrecords.AcademicRecordDeleteForm"),
			),
		).
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware).
			WithQueryPatcher("academicrecords.preload_student_user", views.QueryPatcherPreload("Student.User")).
			WithQueryPatcher("academicrecords.preload_semester", views.QueryPatcherPreload("Semester")).
			WithQueryPatcher("academicrecords.scope_by_role", AcademicRecordScopeByRole),
	)

	// Selection view
	lago.RegistryView.Register("academicrecords.SelectView",
		views.ListView[AcademicRecord]("academicrecords")(
			lago.GetPageView("academicrecords.AcademicRecordSelectionTable"),
		).
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware).
			WithQueryPatcher("academicrecords.preload_student_user", views.QueryPatcherPreload("Student.User")).
			WithQueryPatcher("academicrecords.preload_semester", views.QueryPatcherPreload("Semester")).
			WithQueryPatcher("academicrecords.scope_by_role", AcademicRecordScopeByRole).
			WithQueryPatcher("academicrecords.filter_env_semester", academicRecordsListSemesterEnvQueryPatcher),
	)
}
