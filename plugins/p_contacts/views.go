package p_contacts

import (
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_users"
	"github.com/lariv-in/lago/registry"
	"github.com/lariv-in/lago/views"
)

func init() {
	lago.RegistryView.Register("contacts.ListView",
		lago.GetPageView("contacts.ContactTable").
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware{}).
			WithMiddleware("contacts.list", views.MiddlewareList[Contact]{
				Key: getters.Static("contacts"),
				QueryPatchers: views.QueryPatchers[Contact]{
					{Key: "contacts.order_name", Value: views.QueryPatcherOrderBy[Contact]{Order: "name ASC"}},
				},
			}))

	lago.RegistryView.Register("contacts.DetailView",
		lago.GetPageView("contacts.ContactDetail").
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware{}).
			WithMiddleware("contacts.detail", views.MiddlewareDetail[Contact]{
				Key:          getters.Static("contact"),
				PathParamKey: getters.Static("id"),
			}))

	lago.RegistryView.Register("contacts.CreateView",
		lago.GetPageView("contacts.ContactCreateForm").
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware{}).
			WithMiddleware("contacts.create", views.MiddlewareCreate[Contact]{
				SuccessURL: lago.RoutePath("contacts.DetailRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("$id")),
				}),
			}))

	lago.RegistryView.Register("contacts.UpdateView",
		lago.GetPageView("contacts.ContactUpdateForm").
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware{}).
			WithMiddleware("contacts.detail", views.MiddlewareDetail[Contact]{
				Key:          getters.Static("contact"),
				PathParamKey: getters.Static("id"),
			}).
			WithMiddleware("contacts.update", views.MiddlewareUpdate[Contact]{
				Key: getters.Static("contact"),
				SuccessURL: lago.RoutePath("contacts.DetailRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("contact.ID")),
				}),
			}))

	lago.RegistryView.Register("contacts.DeleteView",
		lago.GetPageView("contacts.ContactDeleteForm").
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware{}).
			WithMiddleware("contacts.detail", views.MiddlewareDetail[Contact]{
				Key:          getters.Static("contact"),
				PathParamKey: getters.Static("id"),
			}).
			WithMiddleware("contacts.delete", views.MiddlewareDelete[Contact]{
				Key:        getters.Static("contact"),
				SuccessURL: lago.RoutePath("contacts.DefaultRoute", nil),
			}))

	lago.RegistryView.Register("contacts.SelectView",
		lago.GetPageView("contacts.ContactSelectionTable").
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware{}).
			WithMiddleware("contacts.select", views.MiddlewareList[Contact]{
				Key: getters.Static("contacts"),
				QueryPatchers: views.QueryPatchers[Contact]{
					registry.Pair[string, views.QueryPatcher[Contact]]{
						Key:   "contacts.order_name",
						Value: views.QueryPatcherOrderBy[Contact]{Order: "name ASC"},
					},
				},
			}))
}
