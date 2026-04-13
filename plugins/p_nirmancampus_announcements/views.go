package p_nirmancampus_announcements

import (
	"net/http"
	"time"

	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_users"
	"github.com/lariv-in/lago/registry"
	"github.com/lariv-in/lago/views"
)

// announcementsAdminRoleLayer limits create/update/delete to the admin role;
// superusers are always allowed (see p_users.RoleAuthorizationLayer).
var announcementsAdminRoleLayer = p_users.RoleAuthorizationLayer{Roles: []string{"admin"}}

type announcementsFormCreatedByPatcher struct{}

func (announcementsFormCreatedByPatcher) Patch(_ views.View, r *http.Request, formData map[string]any, formErrors map[string]error) (map[string]any, map[string]error) {
	user := p_users.UserFromContext(r.Context(), "announcementsFormCreatedByPatcher")
	id := user.ID
	formData["CreatedByID"] = &id
	return formData, formErrors
}

type announcementsFormExpiryAtPointerPatcher struct{}

func (announcementsFormExpiryAtPointerPatcher) Patch(_ views.View, _ *http.Request, formData map[string]any, formErrors map[string]error) (map[string]any, map[string]error) {
	raw, ok := formData["ExpiryAt"]
	if !ok {
		return formData, formErrors
	}
	if raw == nil {
		formData["ExpiryAt"] = nil
		return formData, formErrors
	}

	switch typed := raw.(type) {
	case time.Time:
		if typed.IsZero() {
			formData["ExpiryAt"] = nil
			return formData, formErrors
		}
		tmp := typed
		formData["ExpiryAt"] = &tmp
	case *time.Time:
		// Keep as-is (nil allowed).
	default:
		// Leave unknown types alone; decoding will surface errors if needed.
	}
	return formData, formErrors
}

func init() {
	// List view.
	lago.RegistryView.Register("announcements.ListView",
		lago.GetPageView("announcements.AnnouncementTable").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("announcements.list", views.LayerList[Announcement]{
				Key: getters.Static("announcements"),
				QueryPatchers: views.QueryPatchers[Announcement]{
					registry.Pair[string, views.QueryPatcher[Announcement]]{Key: "announcements.order_release_at", Value: announcementsOrderReleaseAtQueryPatcher},
					registry.Pair[string, views.QueryPatcher[Announcement]]{Key: "announcements.scope_by_role", Value: AnnouncementScopeByRole},
				},
			}))

	// Detail view.
	lago.RegistryView.Register("announcements.DetailView",
		lago.GetPageView("announcements.AnnouncementDetail").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("announcements.detail", views.LayerDetail[Announcement]{
				Key:          getters.Static("announcement"),
				PathParamKey: getters.Static("id"),
				QueryPatchers: views.QueryPatchers[Announcement]{
					registry.Pair[string, views.QueryPatcher[Announcement]]{Key: "announcements.scope_by_role", Value: AnnouncementScopeByRole},
				},
			}))

	// Create view.
	lago.RegistryView.Register("announcements.CreateView",
		lago.GetPageView("announcements.AnnouncementCreateForm").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("announcements.admin_role", announcementsAdminRoleLayer).
			WithLayer("announcements.create", views.LayerCreate[Announcement]{
				SuccessURL: lago.RoutePath("announcements.DetailRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("$id")),
				}),
				FormPatchers: views.FormPatchers{
					registry.Pair[string, views.FormPatcher]{Key: "announcements.form_created_by", Value: announcementsFormCreatedByPatcher{}},
					registry.Pair[string, views.FormPatcher]{Key: "announcements.form_expiry_at", Value: announcementsFormExpiryAtPointerPatcher{}},
				},
			}))

	// Update view.
	lago.RegistryView.Register("announcements.UpdateView",
		lago.GetPageView("announcements.AnnouncementUpdateForm").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("announcements.admin_role", announcementsAdminRoleLayer).
			WithLayer("announcements.detail", views.LayerDetail[Announcement]{
				Key:          getters.Static("announcement"),
				PathParamKey: getters.Static("id"),
				QueryPatchers: views.QueryPatchers[Announcement]{
					registry.Pair[string, views.QueryPatcher[Announcement]]{Key: "announcements.scope_by_role", Value: AnnouncementScopeByRole},
				},
			}).
			WithLayer("announcements.update", views.LayerUpdate[Announcement]{
				Key: getters.Static("announcement"),
				SuccessURL: lago.RoutePath("announcements.DetailRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("announcement.ID")),
				}),
				QueryPatchers: views.QueryPatchers[Announcement]{
					registry.Pair[string, views.QueryPatcher[Announcement]]{Key: "announcements.scope_by_role", Value: AnnouncementScopeByRole},
				},
				FormPatchers: views.FormPatchers{
					registry.Pair[string, views.FormPatcher]{Key: "announcements.form_expiry_at", Value: announcementsFormExpiryAtPointerPatcher{}},
				},
			}))

	// Delete view.
	lago.RegistryView.Register("announcements.DeleteView",
		lago.GetPageView("announcements.AnnouncementDeleteForm").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("announcements.admin_role", announcementsAdminRoleLayer).
			WithLayer("announcements.detail", views.LayerDetail[Announcement]{
				Key:          getters.Static("announcement"),
				PathParamKey: getters.Static("id"),
				QueryPatchers: views.QueryPatchers[Announcement]{
					registry.Pair[string, views.QueryPatcher[Announcement]]{Key: "announcements.scope_by_role", Value: AnnouncementScopeByRole},
				},
			}).
			WithLayer("announcements.delete", views.LayerDelete[Announcement]{
				Key:        getters.Static("announcement"),
				SuccessURL: lago.RoutePath("announcements.DefaultRoute", nil),
				QueryPatchers: views.QueryPatchers[Announcement]{
					registry.Pair[string, views.QueryPatcher[Announcement]]{Key: "announcements.scope_by_role", Value: AnnouncementScopeByRole},
				},
			}))

	// Selection view.
	lago.RegistryView.Register("announcements.SelectView",
		lago.GetPageView("announcements.AnnouncementSelectionTable").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("announcements.select", views.LayerList[Announcement]{
				Key: getters.Static("announcements"),
				QueryPatchers: views.QueryPatchers[Announcement]{
					registry.Pair[string, views.QueryPatcher[Announcement]]{Key: "announcements.order_release_at", Value: announcementsOrderReleaseAtQueryPatcher},
					registry.Pair[string, views.QueryPatcher[Announcement]]{Key: "announcements.scope_by_role", Value: AnnouncementScopeByRole},
				},
			}))
}
