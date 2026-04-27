package p_teachers

import (
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_users"
	"github.com/lariv-in/lago/views"
)

func init() {
	lago.RegistryView.Register("teachers.ListView",
		lago.GetPageView("teachers.TeacherTable").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("teachers.list", views.LayerList[Teacher]{Key: getters.Static("teachers")}))

	lago.RegistryView.Register("teachers.DetailView",
		lago.GetPageView("teachers.TeacherDetail").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("teachers.detail", views.LayerDetail[Teacher]{Key: getters.Static("teacher"), PathParamKey: getters.Static("id")}))

	lago.RegistryView.Register("teachers.CreateView",
		lago.GetPageView("teachers.TeacherCreateForm").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("teachers.create", views.LayerCreate[Teacher]{
				SuccessURL: lago.RoutePath("teachers.DetailRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("$id"))}),
			}))

	lago.RegistryView.Register("teachers.UpdateView",
		lago.GetPageView("teachers.TeacherUpdateForm").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("teachers.detail", views.LayerDetail[Teacher]{Key: getters.Static("teacher"), PathParamKey: getters.Static("id")}).
			WithLayer("teachers.update", views.LayerUpdate[Teacher]{
				Key:        getters.Static("teacher"),
				SuccessURL: lago.RoutePath("teachers.DetailRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("teacher.ID"))}),
			}))

	lago.RegistryView.Register("teachers.DeleteView",
		lago.GetPageView("teachers.TeacherDeleteForm").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("teachers.detail", views.LayerDetail[Teacher]{Key: getters.Static("teacher"), PathParamKey: getters.Static("id")}).
			WithLayer("teachers.delete", views.LayerDelete[Teacher]{Key: getters.Static("teacher"), SuccessURL: lago.RoutePath("teachers.DefaultRoute", nil)}))

	lago.RegistryView.Register("teachers.SelectView",
		lago.GetPageView("teachers.TeacherSelectionTable").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("teachers.select", views.LayerList[Teacher]{Key: getters.Static("teachers")}))

	lago.RegistryView.Register("teachers.MultiSelectView",
		lago.GetPageView("teachers.TeacherMultiSelectionTable").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("teachers.multiselect", views.LayerList[Teacher]{Key: getters.Static("teachers")}))
}
