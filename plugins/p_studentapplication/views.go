package p_studentapplication

import (
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/p_users"
	"github.com/lariv-in/lago/views"
)

func init() {
	lago.RegistryView.Register("studentapplications.ListView",
		views.ListView[StudentApplication]("studentapplications")(
			lago.GetPageView("studentapplications.ApplicationTable")).
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware).
			WithQueryPatcher("studentapplications.preload_program", views.QueryPatcherPreload("Program")))

	lago.RegistryView.Register("studentapplications.DetailView",
		views.DetailView[StudentApplication]("studentapplication")(
			lago.GetPageView("studentapplications.ApplicationDetail")).
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware).
			WithQueryPatcher("studentapplications.preload_program", views.QueryPatcherPreload("Program")))

	lago.RegistryView.Register("studentapplications.CreateView",
		views.CreateView[StudentApplication](lago.GetterRoutePath("studentapplications.DetailRoute", map[string]getters.Getter[any]{"id": getters.GetterAny(getters.GetterKey[uint]("$id"))}))(
			lago.GetPageView("studentapplications.ApplicationCreateForm")).
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware))

	lago.RegistryView.Register("studentapplications.UpdateView",
		views.DetailView[StudentApplication]("studentapplication")(
			views.UpdateView[StudentApplication](lago.GetterRoutePath("studentapplications.DetailRoute", map[string]getters.Getter[any]{"id": getters.GetterAny(getters.GetterKey[uint]("$id"))}))(
				lago.GetPageView("studentapplications.ApplicationUpdateForm"))).
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware).
			WithQueryPatcher("studentapplications.preload_program", views.QueryPatcherPreload("Program")))

	lago.RegistryView.Register("studentapplications.DeleteView",
		views.DetailView[StudentApplication]("studentapplication")(
			views.DeleteView[StudentApplication](lago.GetterRoutePath("studentapplications.DefaultRoute", nil))(
				lago.GetPageView("studentapplications.ApplicationDeleteForm"))).
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware).
			WithQueryPatcher("studentapplications.preload_program", views.QueryPatcherPreload("Program")))
}
