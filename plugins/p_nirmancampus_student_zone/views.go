package p_nirmancampus_student_zone

import (
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/p_users"
	"github.com/lariv-in/lago/views"
)

var roleMiddleware = p_users.RoleAuthorizationMiddleware([]string{"nirmancampus_admin"})

func init() {
	// --- Section views ---

	lago.RegistryView.Register("student_zone.SectionListView",
		views.ListView[StudentZoneSection]("sections")(
			lago.GetPageView("student_zone.SectionTable"),
		).
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware).
			WithMiddleware("student_zone.role", roleMiddleware).
			WithQueryPatcher("student_zone.order", views.QueryPatcherOrderBy("\"order\" ASC")))

	lago.RegistryView.Register("student_zone.SectionDetailView",
		views.DetailView[StudentZoneSection]("section")(
			lago.GetPageView("student_zone.SectionDetail"),
		).
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware).
			WithMiddleware("student_zone.role", roleMiddleware))

	lago.RegistryView.Register("student_zone.SectionCreateView",
		views.CreateView[StudentZoneSection](
			lago.GetterRoutePath("student_zone.SectionDetailRoute", map[string]getters.Getter[any]{
				"id": getters.GetterAny(getters.GetterKey[uint]("$id")),
			}),
		)(
			lago.GetPageView("student_zone.SectionCreateForm"),
		).
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware).
			WithMiddleware("student_zone.role", roleMiddleware))

	lago.RegistryView.Register("student_zone.SectionUpdateView",
		views.DetailView[StudentZoneSection]("section")(
			views.UpdateView[StudentZoneSection](
				lago.GetterRoutePath("student_zone.SectionDetailRoute", map[string]getters.Getter[any]{
					"id": getters.GetterAny(getters.GetterKey[uint]("$id")),
				}),
			)(
				lago.GetPageView("student_zone.SectionUpdateForm"),
			),
		).
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware).
			WithMiddleware("student_zone.role", roleMiddleware))

	lago.RegistryView.Register("student_zone.SectionDeleteView",
		views.DetailView[StudentZoneSection]("section")(
			views.DeleteView[StudentZoneSection](lago.GetterRoutePath("student_zone.DefaultRoute", nil))(
				lago.GetPageView("student_zone.SectionDeleteForm"),
			),
		).
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware).
			WithMiddleware("student_zone.role", roleMiddleware))

	lago.RegistryView.Register("student_zone.SectionSelectView",
		views.ListView[StudentZoneSection]("sections")(
			lago.GetPageView("student_zone.SectionSelectionTable"),
		).
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware).
			WithMiddleware("student_zone.role", roleMiddleware).
			WithQueryPatcher("student_zone.order", views.QueryPatcherOrderBy("\"order\" ASC")))

	// --- Item views ---

	lago.RegistryView.Register("student_zone.ItemListView",
		views.ListView[StudentZoneItem]("items")(
			lago.GetPageView("student_zone.ItemTable"),
		).
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware).
			WithMiddleware("student_zone.role", roleMiddleware).
			WithQueryPatcher("student_zone.preload_section", views.QueryPatcherPreload("StudentZoneSection")).
			WithQueryPatcher("student_zone.preload_file", views.QueryPatcherPreload("File")))

	lago.RegistryView.Register("student_zone.ItemDetailView",
		views.DetailView[StudentZoneItem]("item")(
			lago.GetPageView("student_zone.ItemDetail"),
		).
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware).
			WithMiddleware("student_zone.role", roleMiddleware).
			WithQueryPatcher("student_zone.preload_section", views.QueryPatcherPreload("StudentZoneSection")).
			WithQueryPatcher("student_zone.preload_file", views.QueryPatcherPreload("File")))

	lago.RegistryView.Register("student_zone.ItemCreateView",
		views.CreateView[StudentZoneItem](
			lago.GetterRoutePath("student_zone.ItemDetailRoute", map[string]getters.Getter[any]{
				"id": getters.GetterAny(getters.GetterKey[uint]("$id")),
			}),
		)(
			lago.GetPageView("student_zone.ItemCreateForm"),
		).
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware).
			WithMiddleware("student_zone.role", roleMiddleware))

	lago.RegistryView.Register("student_zone.ItemUpdateView",
		views.DetailView[StudentZoneItem]("item")(
			views.UpdateView[StudentZoneItem](
				lago.GetterRoutePath("student_zone.ItemDetailRoute", map[string]getters.Getter[any]{
					"id": getters.GetterAny(getters.GetterKey[uint]("$id")),
				}),
			)(
				lago.GetPageView("student_zone.ItemUpdateForm"),
			),
		).
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware).
			WithMiddleware("student_zone.role", roleMiddleware).
			WithQueryPatcher("student_zone.preload_section", views.QueryPatcherPreload("StudentZoneSection")).
			WithQueryPatcher("student_zone.preload_file", views.QueryPatcherPreload("File")))

	lago.RegistryView.Register("student_zone.ItemDeleteView",
		views.DetailView[StudentZoneItem]("item")(
			views.DeleteView[StudentZoneItem](lago.GetterRoutePath("student_zone.ItemListRoute", nil))(
				lago.GetPageView("student_zone.ItemDeleteForm"),
			),
		).
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware).
			WithMiddleware("student_zone.role", roleMiddleware))
}
