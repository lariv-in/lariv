package p_nirmancampus_announcements

import (
	"net/http"
	"time"

	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_users"
	"github.com/lariv-in/lago/views"
)

// announcementsAdminRoleMiddleware limits create/update/delete to the admin role;
// superusers are always allowed (see p_users.RoleAuthorizationMiddleware).
var announcementsAdminRoleMiddleware = p_users.RoleAuthorizationMiddleware([]string{"admin"})

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
			WithQueryPatcher("announcements.order_release_at", announcementsOrderReleaseAtQueryPatcher).
			WithQueryPatcher("announcements.scope_by_role", AnnouncementScopeByRole))

	// Detail view.
	lago.RegistryView.Register("announcements.DetailView",
		views.DetailView[Announcement]("announcement", "id")(
			lago.GetPageView("announcements.AnnouncementDetail"),
		).
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware).
			WithQueryPatcher("announcements.scope_by_role", AnnouncementScopeByRole))

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
			WithMiddleware("announcements.admin_role", announcementsAdminRoleMiddleware).
			WithFormPatcher("announcements.form", announcementsFormCreatedByPatcher).
			WithFormPatcher("announcements.form", announcementsFormExpiryAtPointerPatcher))

	// Update view.
	lago.RegistryView.Register("announcements.UpdateView",
		views.DetailView[Announcement]("announcement", "id")(
			views.UpdateView[Announcement]("id",
				lago.GetterRoutePath("announcements.DetailRoute", map[string]getters.Getter[any]{
					"id": getters.GetterAny(getters.GetterKey[uint]("$id")),
				}),
			)(
				lago.GetPageView("announcements.AnnouncementUpdateForm"),
			),
		).
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware).
			WithMiddleware("announcements.admin_role", announcementsAdminRoleMiddleware).
			WithQueryPatcher("announcements.scope_by_role", AnnouncementScopeByRole).
			WithFormPatcher("announcements.form", announcementsFormExpiryAtPointerPatcher))

	// Delete view.
	lago.RegistryView.Register("announcements.DeleteView",
		views.DetailView[Announcement]("announcement", "id")(
			views.DeleteView[Announcement]("id", lago.GetterRoutePath("announcements.DefaultRoute", nil))(
				lago.GetPageView("announcements.AnnouncementDeleteForm"),
			),
		).
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware).
			WithMiddleware("announcements.admin_role", announcementsAdminRoleMiddleware).
			WithQueryPatcher("announcements.scope_by_role", AnnouncementScopeByRole))

	// Selection view.
	lago.RegistryView.Register("announcements.SelectView",
		views.ListView[Announcement]("announcements")(
			lago.GetPageView("announcements.AnnouncementSelectionTable"),
		).
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware).
			WithQueryPatcher("announcements.order_release_at", announcementsOrderReleaseAtQueryPatcher).
			WithQueryPatcher("announcements.scope_by_role", AnnouncementScopeByRole))
}
