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
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("contacts.list", views.LayerList[Contact]{
				Key: getters.Static("contacts"),
				QueryPatchers: views.QueryPatchers[Contact]{
					{Key: "contacts.order_name", Value: views.QueryPatcherOrderBy[Contact]{Order: "name ASC"}},
				},
			}))

	lago.RegistryView.Register("contacts.DetailView",
		lago.GetPageView("contacts.ContactDetail").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("contacts.detail", views.LayerDetail[Contact]{
				Key:          getters.Static("contact"),
				PathParamKey: getters.Static("id"),
			}))

	lago.RegistryView.Register("contacts.CreateView",
		lago.GetPageView("contacts.ContactCreateForm").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("contacts.create", views.LayerCreate[Contact]{
				SuccessURL: lago.RoutePath("contacts.DetailRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("$id")),
				}),
			}))

	lago.RegistryView.Register("contacts.UpdateView",
		lago.GetPageView("contacts.ContactUpdateForm").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("contacts.detail", views.LayerDetail[Contact]{
				Key:          getters.Static("contact"),
				PathParamKey: getters.Static("id"),
			}).
			WithLayer("contacts.update", views.LayerUpdate[Contact]{
				Key: getters.Static("contact"),
				SuccessURL: lago.RoutePath("contacts.DetailRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("contact.ID")),
				}),
			}))

	lago.RegistryView.Register("contacts.DeleteView",
		lago.GetPageView("contacts.ContactDeleteForm").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("contacts.detail", views.LayerDetail[Contact]{
				Key:          getters.Static("contact"),
				PathParamKey: getters.Static("id"),
			}).
			WithLayer("contacts.delete", views.LayerDelete[Contact]{
				Key:        getters.Static("contact"),
				SuccessURL: lago.RoutePath("contacts.DefaultRoute", nil),
			}))

	lago.RegistryView.Register("contacts.SelectView",
		lago.GetPageView("contacts.ContactSelectionTable").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("contacts.select", views.LayerList[Contact]{
				Key: getters.Static("contacts"),
				QueryPatchers: views.QueryPatchers[Contact]{
					registry.Pair[string, views.QueryPatcher[Contact]]{
						Key:   "contacts.order_name",
						Value: views.QueryPatcherOrderBy[Contact]{Order: "name ASC"},
					},
				},
			}))
}
