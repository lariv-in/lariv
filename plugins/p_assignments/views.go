package p_assignments

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

// assignmentsListSemesterEnvQueryPatcher scopes the list view to the semester
// selected in the environment cookie (components.Environment[uint] semester key).
func assignmentsListSemesterEnvQueryPatcher(_ *views.View, r *http.Request, query *gorm.DB) *gorm.DB {
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

func assignmentsOrderDueQueryPatcher(_ *views.View, r *http.Request, query *gorm.DB) *gorm.DB {
	if r.URL.Query().Get("sort") != "" {
		return query
	}
	return query.Order("due ASC")
}

func init() {
	lago.RegistryView.Register("assignments.ListView",
		views.ListView[Assignment]("assignments")(
			lago.GetPageView("assignments.AssignmentTable"),
		).
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware).
			WithQueryPatcher("assignments.preload_semester", views.QueryPatcherPreload("Semester")).
			WithQueryPatcher("assignments.order_due", assignmentsOrderDueQueryPatcher).
			WithQueryPatcher("assignments.filter_env_semester", assignmentsListSemesterEnvQueryPatcher))

	lago.RegistryView.Register("assignments.DetailView",
		views.DetailView[Assignment]("assignment")(
			lago.GetPageView("assignments.AssignmentDetail"),
		).
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware).
			WithQueryPatcher("assignments.preload_semester", views.QueryPatcherPreload("Semester")).
			WithQueryPatcher("assignments.preload_assets", views.QueryPatcherPreload("Assets")))

	lago.RegistryView.Register("assignments.CreateView",
		views.CreateView[Assignment](
			lago.GetterRoutePath("assignments.DetailRoute", map[string]getters.Getter[any]{
				"id": getters.GetterAny(getters.GetterKey[uint]("$id")),
			}),
		)(
			lago.GetPageView("assignments.AssignmentCreateForm"),
		).
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware))

	lago.RegistryView.Register("assignments.UpdateView",
		views.DetailView[Assignment]("assignment")(
			views.UpdateView[Assignment](
				lago.GetterRoutePath("assignments.DetailRoute", map[string]getters.Getter[any]{
					"id": getters.GetterAny(getters.GetterKey[uint]("$id")),
				}),
			)(
				lago.GetPageView("assignments.AssignmentUpdateForm"),
			),
		).
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware).
			WithQueryPatcher("assignments.preload_semester", views.QueryPatcherPreload("Semester")).
			WithQueryPatcher("assignments.preload_assets", views.QueryPatcherPreload("Assets")))

	lago.RegistryView.Register("assignments.DeleteView",
		views.DetailView[Assignment]("assignment")(
			views.DeleteView[Assignment](
				lago.GetterRoutePath("assignments.DefaultRoute", nil),
			)(
				lago.GetPageView("assignments.AssignmentDeleteForm"),
			),
		).
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware).
			WithQueryPatcher("assignments.preload_semester", views.QueryPatcherPreload("Semester")).
			WithQueryPatcher("assignments.preload_assets", views.QueryPatcherPreload("Assets")))

	lago.RegistryView.Register("assignments.SelectView",
		views.ListView[Assignment]("assignments")(
			lago.GetPageView("assignments.AssignmentSelectionTable"),
		).
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware).
			WithQueryPatcher("assignments.preload_semester", views.QueryPatcherPreload("Semester")).
			WithQueryPatcher("assignments.order_due", assignmentsOrderDueQueryPatcher).
			WithQueryPatcher("assignments.filter_env_semester", assignmentsListSemesterEnvQueryPatcher))
}
