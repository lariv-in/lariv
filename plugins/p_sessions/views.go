package p_sessions

import (
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_users"
	"github.com/lariv-in/lago/registry"
	"github.com/lariv-in/lago/views"
)

var classSessionPreload = views.QueryPatchers[ClassSession]{
	registry.Pair[string, views.QueryPatcher[ClassSession]]{Key: "sessions.fk", Value: views.QueryPatcherPreload[ClassSession]{Fields: []string{"Semester", "Course"}}},
}

func init() {
	lago.RegistryView.Register("sessions.ListView",
		lago.GetPageView("sessions.ClassSessionTable").WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("sessions.list", views.LayerList[ClassSession]{Key: getters.Static("class_sessions"), QueryPatchers: classSessionPreload}))
	lago.RegistryView.Register("sessions.DetailView",
		lago.GetPageView("sessions.ClassSessionDetail").WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("sessions.detail", views.LayerDetail[ClassSession]{Key: getters.Static("class_session"), PathParamKey: getters.Static("id"), QueryPatchers: classSessionPreload}))
	lago.RegistryView.Register("sessions.CreateView",
		lago.GetPageView("sessions.ClassSessionCreateForm").WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("sessions.create", views.LayerCreate[ClassSession]{SuccessURL: lago.RoutePath("sessions.DetailRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("$id"))})}))
	lago.RegistryView.Register("sessions.UpdateView",
		lago.GetPageView("sessions.ClassSessionUpdateForm").WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("sessions.detail", views.LayerDetail[ClassSession]{Key: getters.Static("class_session"), PathParamKey: getters.Static("id"), QueryPatchers: classSessionPreload}).
			WithLayer("sessions.update", views.LayerUpdate[ClassSession]{Key: getters.Static("class_session"), SuccessURL: lago.RoutePath("sessions.DetailRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("class_session.ID"))})}))
	lago.RegistryView.Register("sessions.DeleteView",
		lago.GetPageView("sessions.ClassSessionDeleteForm").WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("sessions.detail", views.LayerDetail[ClassSession]{Key: getters.Static("class_session"), PathParamKey: getters.Static("id"), QueryPatchers: classSessionPreload}).
			WithLayer("sessions.delete", views.LayerDelete[ClassSession]{Key: getters.Static("class_session"), SuccessURL: lago.RoutePath("sessions.DefaultRoute", nil)}))
	lago.RegistryView.Register("sessions.SelectView",
		lago.GetPageView("sessions.ClassSessionSelectionTable").WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("sessions.select", views.LayerList[ClassSession]{Key: getters.Static("class_sessions"), QueryPatchers: classSessionPreload}))
}
