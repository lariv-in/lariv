package p_announcements

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/lariv-in/getters"
	"github.com/lariv-in/lago"
	"github.com/lariv-in/p_semesters"
	"github.com/lariv-in/p_users"
	"github.com/lariv-in/views"
	"gorm.io/gorm"
)

func parseSemesterEnvID(raw string) (uint, bool) {
	// components.Environment stores values as "ID:Name".
	parts := strings.SplitN(raw, ":", 2)
	if len(parts) == 0 || strings.TrimSpace(parts[0]) == "" {
		return 0, false
	}
	id, err := strconv.ParseUint(parts[0], 10, 64)
	if err != nil || id == 0 {
		return 0, false
	}
	return uint(id), true
}

// semesterEnvironmentDefault returns the environment option "ID:Name" for the
// semester whose [Start, End] interval contains now (inclusive), or ("", false) if none.
func semesterEnvironmentDefault(db *gorm.DB, now time.Time) (string, bool) {
	var sem p_semesters.Semester
	err := db.Model(&p_semesters.Semester{}).
		Where("start <= ? AND end >= ?", now, now).
		Order("start ASC").
		First(&sem).Error
	if err != nil {
		return "", false
	}
	return fmt.Sprintf("%d:%s", sem.ID, sem.Name), true
}

// announcementsListSemesterEnvQueryPatcher scopes the list view to the semester
// selected in the environment cookie (components.Environment).
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
		var found bool
		raw, found = semesterEnvironmentDefault(db, time.Now())
		if !found {
			return query
		}
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
