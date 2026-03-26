package p_nirmancampus_website

import (
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_users"
	"github.com/lariv-in/lago/views"
)

var importantLinksAdminRoleMiddleware = p_users.RoleAuthorizationMiddleware([]string{"nirmancampus_admin"})

func init() {
	// --- List ---
	lago.RegistryView.Register("nirmancampus_website.ImportantLinksListView",
		views.ListView[ImportantLink]("links")(
			lago.GetPageView("nirmancampus_website.ImportantLinksTable"),
		).
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware).
			WithMiddleware("important_links_admin.role", importantLinksAdminRoleMiddleware).
			WithQueryPatcher("important_links_admin.order", views.QueryPatcherOrderBy("\"order\" ASC")).
			WithQueryPatcher("important_links_admin.preload_file", views.QueryPatcherPreload("File")))

	// --- Detail ---
	lago.RegistryView.Register("nirmancampus_website.ImportantLinksDetailView",
		views.DetailView[ImportantLink]("link")(
			lago.GetPageView("nirmancampus_website.ImportantLinksDetail"),
		).
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware).
			WithMiddleware("important_links_admin.role", importantLinksAdminRoleMiddleware).
			WithQueryPatcher("important_links_admin.preload_file", views.QueryPatcherPreload("File")))

	// --- Create ---
	lago.RegistryView.Register("nirmancampus_website.ImportantLinksCreateView",
		views.CreateView[ImportantLink](
			lago.GetterRoutePath("nirmancampus_website.ImportantLinksDetailRoute", map[string]getters.Getter[any]{
				"id": getters.GetterAny(getters.GetterKey[uint]("$id")),
			}),
		)(
			lago.GetPageView("nirmancampus_website.ImportantLinksCreateForm"),
		).
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware).
			WithMiddleware("important_links_admin.role", importantLinksAdminRoleMiddleware))

	// --- Import ---
	lago.RegistryView.Register("nirmancampus_website.ImportantLinksImportView",
		views.JsonImport[ImportantLink](
			"ImportFile",
			lago.GetterRoutePath("nirmancampus_website.ImportantLinksDefaultRoute", nil),
		)(
			lago.GetPageView("nirmancampus_website.ImportantLinksImportForm"),
		).
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware).
			WithMiddleware("important_links_admin.role", importantLinksAdminRoleMiddleware))

	// --- Update ---
	lago.RegistryView.Register("nirmancampus_website.ImportantLinksUpdateView",
		views.DetailView[ImportantLink]("link")(
			views.UpdateView[ImportantLink](
				lago.GetterRoutePath("nirmancampus_website.ImportantLinksDetailRoute", map[string]getters.Getter[any]{
					"id": getters.GetterAny(getters.GetterKey[uint]("$id")),
				}),
			)(
				lago.GetPageView("nirmancampus_website.ImportantLinksUpdateForm"),
			),
		).
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware).
			WithMiddleware("important_links_admin.role", importantLinksAdminRoleMiddleware).
			WithQueryPatcher("important_links_admin.preload_file", views.QueryPatcherPreload("File")))

	// --- Delete ---
	lago.RegistryView.Register("nirmancampus_website.ImportantLinksDeleteView",
		views.DetailView[ImportantLink]("link")(
			views.DeleteView[ImportantLink](lago.GetterRoutePath("nirmancampus_website.ImportantLinksDefaultRoute", nil))(
				lago.GetPageView("nirmancampus_website.ImportantLinksDeleteForm"),
			),
		).
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware).
			WithMiddleware("important_links_admin.role", importantLinksAdminRoleMiddleware))
}

