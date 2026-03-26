package p_nirmancampus_courses

import "github.com/lariv-in/lago/lago"

const courseProgramAppURL = "/course-program-mappings/"

func registerCourseProgramRoutes() {
	_ = lago.RegistryRoute.Register("courses.CourseProgramDefaultRoute", lago.Route{
		Path:    courseProgramAppURL,
		Handler: lago.NewDynamicView("courses.CourseProgramListView"),
	})

	_ = lago.RegistryRoute.Register("courses.CourseProgramCreateRoute", lago.Route{
		Path:    courseProgramAppURL + "create/",
		Handler: lago.NewDynamicView("courses.CourseProgramCreateView"),
	})

	_ = lago.RegistryRoute.Register("courses.CourseProgramDetailRoute", lago.Route{
		Path:    courseProgramAppURL + "{id}/",
		Handler: lago.NewDynamicView("courses.CourseProgramDetailView"),
	})

	_ = lago.RegistryRoute.Register("courses.CourseProgramUpdateRoute", lago.Route{
		Path:    courseProgramAppURL + "{id}/edit/",
		Handler: lago.NewDynamicView("courses.CourseProgramUpdateView"),
	})

	_ = lago.RegistryRoute.Register("courses.CourseProgramDeleteRoute", lago.Route{
		Path:    courseProgramAppURL + "{id}/delete/",
		Handler: lago.NewDynamicView("courses.CourseProgramDeleteView"),
	})

	_ = lago.RegistryRoute.Register("courses.CourseProgramSelectRoute", lago.Route{
		Path:    courseProgramAppURL + "select/",
		Handler: lago.NewDynamicView("courses.CourseProgramSelectView"),
	})
}

func init() {
	registerCourseProgramRoutes()
}
