package p_programs

import (
	"github.com/lariv-in/getters"
	"github.com/lariv-in/lago"
	"github.com/lariv-in/p_users"
	"github.com/lariv-in/views"
)

func init() {
	// List view
	lago.RegistryView.Register("programs.ListView",
		views.ListView[Program]("programs")(
			lago.GetPageView("programs.ProgramTable"),
		).
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware))

	// Detail view
	lago.RegistryView.Register("programs.DetailView",
		views.DetailView[Program]("program")(
			lago.GetPageView("programs.ProgramDetail"),
		).
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware))

	// Create view
	lago.RegistryView.Register("programs.CreateView",
		views.CreateView[Program](
			lago.GetterRoutePath("programs.DetailRoute", map[string]getters.Getter[any]{
				"id": getters.GetterAny(getters.GetterKey[uint]("$id")),
			}),
		)(
			lago.GetPageView("programs.ProgramCreateForm"),
		).
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware))

	// Update view
	lago.RegistryView.Register("programs.UpdateView",
		views.DetailView[Program]("program")(
			views.UpdateView[Program](
				lago.GetterRoutePath("programs.DetailRoute", map[string]getters.Getter[any]{
					"id": getters.GetterAny(getters.GetterKey[uint]("$id")),
				}),
			)(
				lago.GetPageView("programs.ProgramUpdateForm"),
			),
		).
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware))

	// Delete view
	lago.RegistryView.Register("programs.DeleteView",
		views.DetailView[Program]("program")(
			views.DeleteView[Program](
				lago.GetterRoutePath("programs.DefaultRoute", nil),
			)(
				lago.GetPageView("programs.ProgramDeleteForm"),
			),
		).
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware))

	// Selection view
	lago.RegistryView.Register("programs.SelectView",
		views.ListView[Program]("programs")(
			lago.GetPageView("programs.ProgramSelectionTable"),
		).
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware))
}

