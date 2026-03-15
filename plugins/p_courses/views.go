package p_courses

import (
	"github.com/lariv-in/getters"
	"github.com/lariv-in/lago"
	"github.com/lariv-in/p_users"
	"github.com/lariv-in/views"
)

var courseModel = Course{}

func init() {
	// List view
	lago.RegistryView.Register("courses.ListView",
		p_users.AuthenticationMiddleware(
			views.ListView[Course]("courses")(
				lago.GetPageView("courses.CourseTable"))))

	// Detail view
	lago.RegistryView.Register("courses.DetailView",
		p_users.AuthenticationMiddleware(
			views.DetailView[Course]("course")(
				lago.GetPageView("courses.CourseDetail"))))

	// Create view
	lago.RegistryView.Register("courses.CreateView",
		p_users.AuthenticationMiddleware(
			views.CreateView[Course](lago.GetterRoutePath("courses.DetailRoute", map[string]getters.Getter{"id": getters.GetterKey("$id")}))(
				lago.GetPageView("courses.CourseCreateForm"))))

	// Update view
	lago.RegistryView.Register("courses.UpdateView",
		p_users.AuthenticationMiddleware(
			views.DetailView[Course]("course")(
				views.UpdateView[Course](lago.GetterRoutePath("courses.DetailRoute", map[string]getters.Getter{"id": getters.GetterKey("$id")}))(
					lago.GetPageView("courses.CourseUpdateForm")))))

	// Delete view
	lago.RegistryView.Register("courses.DeleteView",
		p_users.AuthenticationMiddleware(
			views.DetailView[Course]("course")(
				views.DeleteView[Course](lago.GetterRoutePath("courses.DefaultRoute", nil))(
					lago.GetPageView("courses.CourseDeleteForm")))))

	// Selection views
	lago.RegistryView.Register("courses.SelectView",
		p_users.AuthenticationMiddleware(
			views.ListView[Course]("courses")(
				lago.GetPageView("courses.CourseSelectionTable"))))

	lago.RegistryView.Register("courses.MultiSelectView",
		p_users.AuthenticationMiddleware(
			views.ListView[Course]("courses")(
				lago.GetPageView("courses.CourseMultiSelectionTable"))))
}

