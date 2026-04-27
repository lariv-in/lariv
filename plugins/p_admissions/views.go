package p_admissions

import (
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_users"
	"github.com/lariv-in/lago/registry"
	"github.com/lariv-in/lago/views"
)

var admissionPreload = views.QueryPatchers[AdmissionApplication]{
	registry.Pair[string, views.QueryPatcher[AdmissionApplication]]{Key: "admissions.fk", Value: views.QueryPatcherPreload[AdmissionApplication]{Fields: []string{"Program", "Semester", "LinkedUser"}}},
}

func init() {
	lago.RegistryView.Register("admissions.ListView", lago.GetPageView("admissions.ApplicationTable").WithLayer("users.auth", p_users.AuthenticationLayer{}).WithLayer("admissions.list", views.LayerList[AdmissionApplication]{Key: getters.Static("applications")}))
	lago.RegistryView.Register("admissions.DetailView", lago.GetPageView("admissions.ApplicationDetail").WithLayer("users.auth", p_users.AuthenticationLayer{}).WithLayer("admissions.detail", views.LayerDetail[AdmissionApplication]{Key: getters.Static("application"), PathParamKey: getters.Static("id"), QueryPatchers: admissionPreload}))
	lago.RegistryView.Register("admissions.CreateView", lago.GetPageView("admissions.ApplicationCreateForm").WithLayer("users.auth", p_users.AuthenticationLayer{}).WithLayer("admissions.create", views.LayerCreate[AdmissionApplication]{SuccessURL: lago.RoutePath("admissions.DetailRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("$id"))})}))
	lago.RegistryView.Register("admissions.UpdateView", lago.GetPageView("admissions.ApplicationUpdateForm").WithLayer("users.auth", p_users.AuthenticationLayer{}).WithLayer("admissions.detail", views.LayerDetail[AdmissionApplication]{Key: getters.Static("application"), PathParamKey: getters.Static("id"), QueryPatchers: admissionPreload}).WithLayer("admissions.update", views.LayerUpdate[AdmissionApplication]{Key: getters.Static("application"), SuccessURL: lago.RoutePath("admissions.DetailRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("application.ID"))})}))
	lago.RegistryView.Register("admissions.DeleteView", lago.GetPageView("admissions.ApplicationDeleteForm").WithLayer("users.auth", p_users.AuthenticationLayer{}).WithLayer("admissions.detail", views.LayerDetail[AdmissionApplication]{Key: getters.Static("application"), PathParamKey: getters.Static("id")}).WithLayer("admissions.delete", views.LayerDelete[AdmissionApplication]{Key: getters.Static("application"), SuccessURL: lago.RoutePath("admissions.DefaultRoute", nil)}))
}
