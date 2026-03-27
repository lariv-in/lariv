package p_nirmancampus_studentapplications

import (
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_users"
	"github.com/lariv-in/lago/views"
)

func init() {
	lago.RegistryView.Register("studentapplications.ListView",
		views.ListView[StudentApplication]("studentapplications")(
			lago.GetPageView("studentapplications.ApplicationTable")).
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware).
			WithQueryPatcher("studentapplications.preload_program", views.QueryPatcherPreload("Program")))

	lago.RegistryView.Register("studentapplications.DetailView",
		views.DetailView[StudentApplication]("studentapplication", "id")(
			lago.GetPageView("studentapplications.ApplicationDetail")).
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware).
			WithQueryPatcher("studentapplications.preload_program", views.QueryPatcherPreload("Program")).
			WithQueryPatcher("studentapplications.preload_photo", views.QueryPatcherPreload("Photo")).
			WithQueryPatcher("studentapplications.preload_documents", views.QueryPatcherPreload("Documents")))

	lago.RegistryView.Register("studentapplications.CreateView",
		views.CreateView[StudentApplication](lago.GetterRoutePath("studentapplications.DetailRoute", map[string]getters.Getter[any]{"id": getters.GetterAny(getters.GetterKey[uint]("$id"))}))(
			lago.GetPageView("studentapplications.ApplicationCreateForm")).
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware))

	lago.RegistryView.Register("studentapplications.UpdateView",
		views.DetailView[StudentApplication]("studentapplication", "id")(
			views.UpdateView[StudentApplication]("id", lago.GetterRoutePath("studentapplications.DetailRoute", map[string]getters.Getter[any]{"id": getters.GetterAny(getters.GetterKey[uint]("$id"))}))(
				lago.GetPageView("studentapplications.ApplicationUpdateForm"))).
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware).
			WithQueryPatcher("studentapplications.preload_program", views.QueryPatcherPreload("Program")).
			WithQueryPatcher("studentapplications.preload_photo", views.QueryPatcherPreload("Photo")).
			WithQueryPatcher("studentapplications.preload_documents", views.QueryPatcherPreload("Documents")))

	lago.RegistryView.Register("studentapplications.DeleteView",
		views.DetailView[StudentApplication]("studentapplication", "id")(
			views.DeleteView[StudentApplication]("id", lago.GetterRoutePath("studentapplications.DefaultRoute", nil))(
				lago.GetPageView("studentapplications.ApplicationDeleteForm"))).
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware).
			WithQueryPatcher("studentapplications.preload_program", views.QueryPatcherPreload("Program")))
}
