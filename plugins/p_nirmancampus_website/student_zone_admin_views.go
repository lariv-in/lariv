package p_nirmancampus_website

import (
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_users"
	"github.com/lariv-in/lago/views"
)

var studentZoneAdminRoleMiddleware = p_users.RoleAuthorizationMiddleware([]string{"admin"})

func init() {
	// --- Section views ---

	lago.RegistryView.Register("nirmancampus_website.StudentZoneAdminSectionListView",
		views.ListView[StudentZoneSection]("sections")(
			lago.GetPageView("nirmancampus_website.StudentZoneAdminSectionTable"),
		).
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware).
			WithMiddleware("student_zone_admin.role", studentZoneAdminRoleMiddleware).
			WithQueryPatcher("student_zone_admin.order", views.QueryPatcherOrderBy("\"order\" ASC")))

	lago.RegistryView.Register("nirmancampus_website.StudentZoneAdminSectionDetailView",
		views.DetailView[StudentZoneSection]("section")(
			lago.GetPageView("nirmancampus_website.StudentZoneAdminSectionDetail"),
		).
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware).
			WithMiddleware("student_zone_admin.role", studentZoneAdminRoleMiddleware))

	lago.RegistryView.Register("nirmancampus_website.StudentZoneAdminSectionCreateView",
		views.CreateView[StudentZoneSection](
			lago.GetterRoutePath("nirmancampus_website.StudentZoneAdminSectionDetailRoute", map[string]getters.Getter[any]{
				"id": getters.GetterAny(getters.GetterKey[uint]("$id")),
			}),
		)(
			lago.GetPageView("nirmancampus_website.StudentZoneAdminSectionCreateForm"),
		).
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware).
			WithMiddleware("student_zone_admin.role", studentZoneAdminRoleMiddleware))

	lago.RegistryView.Register("nirmancampus_website.StudentZoneAdminSectionUpdateView",
		views.DetailView[StudentZoneSection]("section")(
			views.UpdateView[StudentZoneSection](
				lago.GetterRoutePath("nirmancampus_website.StudentZoneAdminSectionDetailRoute", map[string]getters.Getter[any]{
					"id": getters.GetterAny(getters.GetterKey[uint]("$id")),
				}),
			)(
				lago.GetPageView("nirmancampus_website.StudentZoneAdminSectionUpdateForm"),
			),
		).
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware).
			WithMiddleware("student_zone_admin.role", studentZoneAdminRoleMiddleware))

	lago.RegistryView.Register("nirmancampus_website.StudentZoneAdminSectionDeleteView",
		views.DetailView[StudentZoneSection]("section")(
			views.DeleteView[StudentZoneSection](lago.GetterRoutePath("nirmancampus_website.StudentZoneAdminDefaultRoute", nil))(
				lago.GetPageView("nirmancampus_website.StudentZoneAdminSectionDeleteForm"),
			),
		).
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware).
			WithMiddleware("student_zone_admin.role", studentZoneAdminRoleMiddleware))

	lago.RegistryView.Register("nirmancampus_website.StudentZoneAdminSectionSelectView",
		views.ListView[StudentZoneSection]("sections")(
			lago.GetPageView("nirmancampus_website.StudentZoneAdminSectionSelectionTable"),
		).
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware).
			WithMiddleware("student_zone_admin.role", studentZoneAdminRoleMiddleware).
			WithQueryPatcher("student_zone_admin.order", views.QueryPatcherOrderBy("\"order\" ASC")))

	// --- Item views ---

	lago.RegistryView.Register("nirmancampus_website.StudentZoneAdminItemListView",
		views.ListView[StudentZoneItem]("items")(
			lago.GetPageView("nirmancampus_website.StudentZoneAdminItemTable"),
		).
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware).
			WithMiddleware("student_zone_admin.role", studentZoneAdminRoleMiddleware).
			WithQueryPatcher("student_zone_admin.preload_section", views.QueryPatcherPreload("StudentZoneSection")).
			WithQueryPatcher("student_zone_admin.preload_file", views.QueryPatcherPreload("File")))

	lago.RegistryView.Register("nirmancampus_website.StudentZoneAdminItemDetailView",
		views.DetailView[StudentZoneItem]("item")(
			lago.GetPageView("nirmancampus_website.StudentZoneAdminItemDetail"),
		).
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware).
			WithMiddleware("student_zone_admin.role", studentZoneAdminRoleMiddleware).
			WithQueryPatcher("student_zone_admin.preload_section", views.QueryPatcherPreload("StudentZoneSection")).
			WithQueryPatcher("student_zone_admin.preload_file", views.QueryPatcherPreload("File")))

	lago.RegistryView.Register("nirmancampus_website.StudentZoneAdminItemCreateView",
		views.CreateView[StudentZoneItem](
			lago.GetterRoutePath("nirmancampus_website.StudentZoneAdminItemDetailRoute", map[string]getters.Getter[any]{
				"id": getters.GetterAny(getters.GetterKey[uint]("$id")),
			}),
		)(
			lago.GetPageView("nirmancampus_website.StudentZoneAdminItemCreateForm"),
		).
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware).
			WithMiddleware("student_zone_admin.role", studentZoneAdminRoleMiddleware))

	lago.RegistryView.Register("nirmancampus_website.StudentZoneAdminItemUpdateView",
		views.DetailView[StudentZoneItem]("item")(
			views.UpdateView[StudentZoneItem](
				lago.GetterRoutePath("nirmancampus_website.StudentZoneAdminItemDetailRoute", map[string]getters.Getter[any]{
					"id": getters.GetterAny(getters.GetterKey[uint]("$id")),
				}),
			)(
				lago.GetPageView("nirmancampus_website.StudentZoneAdminItemUpdateForm"),
			),
		).
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware).
			WithMiddleware("student_zone_admin.role", studentZoneAdminRoleMiddleware).
			WithQueryPatcher("student_zone_admin.preload_section", views.QueryPatcherPreload("StudentZoneSection")).
			WithQueryPatcher("student_zone_admin.preload_file", views.QueryPatcherPreload("File")))

	lago.RegistryView.Register("nirmancampus_website.StudentZoneAdminItemDeleteView",
		views.DetailView[StudentZoneItem]("item")(
			views.DeleteView[StudentZoneItem](lago.GetterRoutePath("nirmancampus_website.StudentZoneAdminItemListRoute", nil))(
				lago.GetPageView("nirmancampus_website.StudentZoneAdminItemDeleteForm"),
			),
		).
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware).
			WithMiddleware("student_zone_admin.role", studentZoneAdminRoleMiddleware))
}

