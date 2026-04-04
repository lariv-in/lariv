package p_nirmancampus_website

import (
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_users"
	"github.com/lariv-in/lago/views"
)

var importantLinksAdminRoleLayer = p_users.RoleAuthorizationLayer{Roles: []string{"admin"}}

func init() {
	// --- List ---
	lago.RegistryView.Register("nirmancampus_website.ImportantLinksListView",
		lago.GetPageView("nirmancampus_website.ImportantLinksTable").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("important_links_admin.role", importantLinksAdminRoleLayer).
			WithLayer("important_links_admin.list", views.LayerList[ImportantLink]{
				Key: getters.Static("links"),
				QueryPatchers: views.QueryPatchers[ImportantLink]{
					{Key: "important_links_admin.order", Value: views.QueryPatcherOrderBy[ImportantLink]{Order: "\"order\" ASC"}},
					{Key: "important_links_admin.preload_file", Value: views.QueryPatcherPreload[ImportantLink]{Field: "File"}},
				},
			}))

	// --- Detail ---
	lago.RegistryView.Register("nirmancampus_website.ImportantLinksDetailView",
		lago.GetPageView("nirmancampus_website.ImportantLinksDetail").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("important_links_admin.role", importantLinksAdminRoleLayer).
			WithLayer("important_links_admin.detail", views.LayerDetail[ImportantLink]{
				Key:          getters.Static("link"),
				PathParamKey: getters.Static("id"),
				QueryPatchers: views.QueryPatchers[ImportantLink]{
					{Key: "important_links_admin.preload_file", Value: views.QueryPatcherPreload[ImportantLink]{Field: "File"}},
				},
			}))

	// --- Create ---
	lago.RegistryView.Register("nirmancampus_website.ImportantLinksCreateView",
		lago.GetPageView("nirmancampus_website.ImportantLinksCreateForm").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("important_links_admin.role", importantLinksAdminRoleLayer).
			WithLayer("important_links_admin.create", views.LayerCreate[ImportantLink]{
				SuccessURL: lago.RoutePath("nirmancampus_website.ImportantLinksDetailRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("$id")),
				}),
			}))

	// --- Import ---
	lago.RegistryView.Register("nirmancampus_website.ImportantLinksImportView",
		lago.GetPageView("nirmancampus_website.ImportantLinksImportForm").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("important_links_admin.role", importantLinksAdminRoleLayer).
			WithLayer("important_links_admin.import", views.LayerJsonImport[ImportantLink]{
				FileField:  "ImportFile",
				SuccessURL: lago.RoutePath("nirmancampus_website.ImportantLinksDefaultRoute", nil),
			}))

	// --- Update ---
	lago.RegistryView.Register("nirmancampus_website.ImportantLinksUpdateView",
		lago.GetPageView("nirmancampus_website.ImportantLinksUpdateForm").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("important_links_admin.role", importantLinksAdminRoleLayer).
			WithLayer("important_links_admin.detail", views.LayerDetail[ImportantLink]{
				Key:          getters.Static("link"),
				PathParamKey: getters.Static("id"),
				QueryPatchers: views.QueryPatchers[ImportantLink]{
					{Key: "important_links_admin.preload_file", Value: views.QueryPatcherPreload[ImportantLink]{Field: "File"}},
				},
			}).
			WithLayer("important_links_admin.update", views.LayerUpdate[ImportantLink]{
				Key: getters.Static("link"),
				SuccessURL: lago.RoutePath("nirmancampus_website.ImportantLinksDetailRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("link.ID")),
				}),
			}))

	// --- Delete ---
	lago.RegistryView.Register("nirmancampus_website.ImportantLinksDeleteView",
		lago.GetPageView("nirmancampus_website.ImportantLinksDeleteForm").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("important_links_admin.role", importantLinksAdminRoleLayer).
			WithLayer("important_links_admin.detail", views.LayerDetail[ImportantLink]{
				Key:          getters.Static("link"),
				PathParamKey: getters.Static("id"),
			}).
			WithLayer("important_links_admin.delete", views.LayerDelete[ImportantLink]{
				Key:        getters.Static("link"),
				SuccessURL: lago.RoutePath("nirmancampus_website.ImportantLinksDefaultRoute", nil),
			}))
}
