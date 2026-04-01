package p_contacts

import (
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_users"
	"github.com/lariv-in/lago/views"
)

func init() {
	lago.RegistryView.Register("contacts.ListView",
		views.ListView[Contact]("contacts")(
			lago.GetPageView("contacts.ContactTable"),
		).
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware).
			WithQueryPatcher("contacts.order_name", views.QueryPatcherOrderBy("name ASC")))

	lago.RegistryView.Register("contacts.DetailView",
		views.DetailView[Contact]("contact", "id")(
			lago.GetPageView("contacts.ContactDetail"),
		).
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware))

	lago.RegistryView.Register("contacts.CreateView",
		views.CreateView[Contact](
			lago.GetterRoutePath("contacts.DetailRoute", map[string]getters.Getter[any]{
				"id": getters.Any(getters.Key[uint]("$id")),
			}),
		)(
			lago.GetPageView("contacts.ContactCreateForm"),
		).
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware))

	lago.RegistryView.Register("contacts.UpdateView",
		views.DetailView[Contact]("contact", "id")(
			views.UpdateView[Contact]("id",
				lago.GetterRoutePath("contacts.DetailRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("$id")),
				}),
			)(
				lago.GetPageView("contacts.ContactUpdateForm"),
			),
		).
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware))

	lago.RegistryView.Register("contacts.DeleteView",
		views.DetailView[Contact]("contact", "id")(
			views.DeleteView[Contact]("id", lago.GetterRoutePath("contacts.DefaultRoute", nil))(
				lago.GetPageView("contacts.ContactDeleteForm"),
			),
		).
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware))

	lago.RegistryView.Register("contacts.SelectView",
		views.ListView[Contact]("contacts")(
			lago.GetPageView("contacts.ContactSelectionTable"),
		).
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware).
			WithQueryPatcher("contacts.order_name", views.QueryPatcherOrderBy("name ASC")))
}
