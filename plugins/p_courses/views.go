package p_courses

import (
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_users"
	"github.com/lariv-in/lago/registry"
	"github.com/lariv-in/lago/views"
)

var courseM2MPreload = views.QueryPatchers[Course]{
	registry.Pair[string, views.QueryPatcher[Course]]{Key: "courses.m2m", Value: views.QueryPatcherPreload[Course]{Fields: []string{"Programs", "Students"}}},
}

func init() {
	lago.RegistryView.Register("courses.ListView",
		lago.GetPageView("courses.CourseTable").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("courses.list", views.LayerList[Course]{Key: getters.Static("courses")}))

	lago.RegistryView.Register("courses.DetailView",
		lago.GetPageView("courses.CourseDetail").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("courses.detail", views.LayerDetail[Course]{
				Key:           getters.Static("course"),
				PathParamKey:  getters.Static("id"),
				QueryPatchers: courseM2MPreload,
			}))

	lago.RegistryView.Register("courses.CreateView",
		lago.GetPageView("courses.CourseCreateForm").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("courses.create", views.LayerCreate[Course]{
				SuccessURL: lago.RoutePath("courses.DetailRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("$id"))}),
			}))

	lago.RegistryView.Register("courses.UpdateView",
		lago.GetPageView("courses.CourseUpdateForm").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("courses.detail", views.LayerDetail[Course]{
				Key:           getters.Static("course"),
				PathParamKey:  getters.Static("id"),
				QueryPatchers: courseM2MPreload,
			}).
			WithLayer("courses.update", views.LayerUpdate[Course]{
				Key:        getters.Static("course"),
				SuccessURL: lago.RoutePath("courses.DetailRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("course.ID"))}),
			}))

	lago.RegistryView.Register("courses.DeleteView",
		lago.GetPageView("courses.CourseDeleteForm").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("courses.detail", views.LayerDetail[Course]{Key: getters.Static("course"), PathParamKey: getters.Static("id")}).
			WithLayer("courses.delete", views.LayerDelete[Course]{Key: getters.Static("course"), SuccessURL: lago.RoutePath("courses.DefaultRoute", nil)}))

	lago.RegistryView.Register("courses.SelectView",
		lago.GetPageView("courses.CourseSelectionTable").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("courses.select", views.LayerList[Course]{Key: getters.Static("courses")}))
}
