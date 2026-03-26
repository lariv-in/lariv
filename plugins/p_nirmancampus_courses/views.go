package p_nirmancampus_courses

import (
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_users"
	"github.com/lariv-in/lago/views"
)

var courseModel = Course{}

func init() {
	// List view
	lago.RegistryView.Register("courses.ListView",
		views.ListView[Course]("courses")(
			lago.GetPageView("courses.CourseTable")).
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware))

	// Detail view
	lago.RegistryView.Register("courses.DetailView",
		views.DetailView[Course]("course")(
			lago.GetPageView("courses.CourseDetail")).
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware))

	// Create view
	lago.RegistryView.Register("courses.CreateView",
		views.CreateView[Course](lago.GetterRoutePath("courses.DetailRoute", map[string]getters.Getter[any]{"id": getters.GetterAny(getters.GetterKey[uint]("$id"))}))(
			lago.GetPageView("courses.CourseCreateForm")).
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware))

	// Update view
	lago.RegistryView.Register("courses.UpdateView",
		views.DetailView[Course]("course")(
			views.UpdateView[Course](lago.GetterRoutePath("courses.DetailRoute", map[string]getters.Getter[any]{"id": getters.GetterAny(getters.GetterKey[uint]("$id"))}))(
				lago.GetPageView("courses.CourseUpdateForm"))).
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware))

	// Delete view
	lago.RegistryView.Register("courses.DeleteView",
		views.DetailView[Course]("course")(
			views.DeleteView[Course](lago.GetterRoutePath("courses.DefaultRoute", nil))(
				lago.GetPageView("courses.CourseDeleteForm"))).
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware))

	// Selection views
	lago.RegistryView.Register("courses.SelectView",
		views.ListView[Course]("courses")(
			lago.GetPageView("courses.CourseSelectionTable")).
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware))

	lago.RegistryView.Register("courses.MultiSelectView",
		views.ListView[Course]("courses")(
			lago.GetPageView("courses.CourseMultiSelectionTable")).
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware))

}
