package p_announcements

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
	// Legacy environment values used "ID:Name"; id is the cookie value prefix.
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

// announcementsListSemesterEnvQueryPatcher scopes the list view to the semester
// selected in the environment cookie (components.Environment[uint] semester key).
func announcementsListSemesterEnvQueryPatcher(_ *views.View, r *http.Request, query *gorm.DB) *gorm.DB {
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

// announcementsOrderReleaseAtQueryPatcher defaults ordering to release_at ASC
// when the request didn't specify sort=.
func announcementsOrderReleaseAtQueryPatcher(_ *views.View, r *http.Request, query *gorm.DB) *gorm.DB {
	if r.URL.Query().Get("sort") != "" {
		return query
	}
	return query.Order("release_at ASC")
}

func announcementsFormCreatedByPatcher(_ *views.View, r *http.Request, formData map[string]any) map[string]any {
	user := r.Context().Value("$user").(p_users.User)
	id := user.ID
	formData["CreatedByID"] = &id
	return formData
}

func announcementsFormExpiryAtPointerPatcher(_ *views.View, _ *http.Request, formData map[string]any) map[string]any {
	raw, ok := formData["ExpiryAt"]
	if !ok {
		return formData
	}
	if raw == nil {
		formData["ExpiryAt"] = nil
		return formData
	}

	switch typed := raw.(type) {
	case time.Time:
		if typed.IsZero() {
			formData["ExpiryAt"] = nil
			return formData
		}
		tmp := typed
		formData["ExpiryAt"] = &tmp
	case *time.Time:
		// Keep as-is (nil allowed).
	default:
		// Leave unknown types alone; decoding will surface errors if needed.
	}
	return formData
}

func init() {
	// List view.
	lago.RegistryView.Register("announcements.ListView",
		views.ListView[Announcement]("announcements")(
			lago.GetPageView("announcements.AnnouncementTable"),
		).
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware).
			WithQueryPatcher("announcements.preload_semester", views.QueryPatcherPreload("Semester")).
			WithQueryPatcher("announcements.order_release_at", announcementsOrderReleaseAtQueryPatcher).
			WithQueryPatcher("announcements.filter_env_semester", announcementsListSemesterEnvQueryPatcher))

	// Detail view.
	lago.RegistryView.Register("announcements.DetailView",
		views.DetailView[Announcement]("announcement")(
			lago.GetPageView("announcements.AnnouncementDetail"),
		).
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware).
			WithQueryPatcher("announcements.preload_semester", views.QueryPatcherPreload("Semester")))

	// Create view.
	lago.RegistryView.Register("announcements.CreateView",
		views.CreateView[Announcement](
			lago.GetterRoutePath("announcements.DetailRoute", map[string]getters.Getter[any]{
				"id": getters.GetterAny(getters.GetterKey[uint]("$id")),
			}),
		)(
			lago.GetPageView("announcements.AnnouncementCreateForm"),
		).
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware).
			WithFormPatcher("announcements.form", announcementsFormCreatedByPatcher).
			WithFormPatcher("announcements.form", announcementsFormExpiryAtPointerPatcher))

	// Update view.
	lago.RegistryView.Register("announcements.UpdateView",
		views.DetailView[Announcement]("announcement")(
			views.UpdateView[Announcement](
				lago.GetterRoutePath("announcements.DetailRoute", map[string]getters.Getter[any]{
					"id": getters.GetterAny(getters.GetterKey[uint]("$id")),
				}),
			)(
				lago.GetPageView("announcements.AnnouncementUpdateForm"),
			),
		).
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware).
			WithFormPatcher("announcements.form", announcementsFormExpiryAtPointerPatcher))

	// Delete view.
	lago.RegistryView.Register("announcements.DeleteView",
		views.DetailView[Announcement]("announcement")(
			views.DeleteView[Announcement](lago.GetterRoutePath("announcements.DefaultRoute", nil))(
				lago.GetPageView("announcements.AnnouncementDeleteForm"),
			),
		).
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware))

	// Selection view.
	lago.RegistryView.Register("announcements.SelectView",
		views.ListView[Announcement]("announcements")(
			lago.GetPageView("announcements.AnnouncementSelectionTable"),
		).
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware).
			WithQueryPatcher("announcements.preload_semester", views.QueryPatcherPreload("Semester")).
			WithQueryPatcher("announcements.order_release_at", announcementsOrderReleaseAtQueryPatcher))
}
