package p_courses

import (
	"github.com/lariv-in/lago"
	"github.com/lariv-in/p_users"
	"github.com/lariv-in/views"
)

var courseModel = Course{}

func init() {
	// List view
	lago.RegistryView.Register("courses.ListView",
		p_users.AuthMiddleware(
			views.ListView(courseModel, "courses")(
				lago.GetPageView("courses.CourseTable"))))

	// Detail view
	lago.RegistryView.Register("courses.DetailView",
		p_users.AuthMiddleware(
			views.DetailView(courseModel, "course")(
				lago.GetPageView("courses.CourseDetail"))))

	// Create view
	lago.RegistryView.Register("courses.CreateView",
		p_users.AuthMiddleware(
			views.CreateView(courseModel, AppUrl+"%v/")(
				lago.GetPageView("courses.CourseCreateForm"))))

	// Update view
	lago.RegistryView.Register("courses.UpdateView",
		p_users.AuthMiddleware(
			views.DetailView(courseModel, "course")(
				views.UpdateView(courseModel, AppUrl+"%v/")(
					lago.GetPageView("courses.CourseUpdateForm")))))

	// Delete view
	lago.RegistryView.Register("courses.DeleteView",
		p_users.AuthMiddleware(
			views.DetailView(courseModel, "course")(
				views.DeleteView(courseModel, AppUrl)(
					lago.GetPageView("courses.CourseDeleteForm")))))

	// Selection views
	lago.RegistryView.Register("courses.SelectView",
		p_users.AuthMiddleware(
			views.ListView(courseModel, "courses")(
				lago.GetPageView("courses.CourseSelectionTable"))))

	lago.RegistryView.Register("courses.MultiSelectView",
		p_users.AuthMiddleware(
			views.ListView(courseModel, "courses")(
				lago.GetPageView("courses.CourseMultiSelectionTable"))))
}

