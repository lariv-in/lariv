package p_semesters

import (
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_users"
	"github.com/lariv-in/lago/views"
)

func init() {
	lago.RegistryView.Register("semesters.ListView", lago.GetPageView("semesters.SemesterTable").WithLayer("users.auth", p_users.AuthenticationLayer{}).WithLayer("semesters.list", views.LayerList[Semester]{Key: getters.Static("semesters")}))
	lago.RegistryView.Register("semesters.DetailView", lago.GetPageView("semesters.SemesterDetail").WithLayer("users.auth", p_users.AuthenticationLayer{}).WithLayer("semesters.detail", views.LayerDetail[Semester]{Key: getters.Static("semester"), PathParamKey: getters.Static("id")}))
	lago.RegistryView.Register("semesters.CreateView", lago.GetPageView("semesters.SemesterCreateForm").WithLayer("users.auth", p_users.AuthenticationLayer{}).WithLayer("semesters.create", views.LayerCreate[Semester]{SuccessURL: lago.RoutePath("semesters.DetailRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("$id"))})}))
	lago.RegistryView.Register("semesters.UpdateView", lago.GetPageView("semesters.SemesterUpdateForm").WithLayer("users.auth", p_users.AuthenticationLayer{}).WithLayer("semesters.detail", views.LayerDetail[Semester]{Key: getters.Static("semester"), PathParamKey: getters.Static("id")}).WithLayer("semesters.update", views.LayerUpdate[Semester]{Key: getters.Static("semester"), SuccessURL: lago.RoutePath("semesters.DetailRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("semester.ID"))})}))
	lago.RegistryView.Register("semesters.DeleteView", lago.GetPageView("semesters.SemesterDeleteForm").WithLayer("users.auth", p_users.AuthenticationLayer{}).WithLayer("semesters.detail", views.LayerDetail[Semester]{Key: getters.Static("semester"), PathParamKey: getters.Static("id")}).WithLayer("semesters.delete", views.LayerDelete[Semester]{Key: getters.Static("semester"), SuccessURL: lago.RoutePath("semesters.DefaultRoute", nil)}))

	lago.RegistryView.Register("semesters.SelectView",
		lago.GetPageView("semesters.SemesterSelectionTable").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("semesters.select", views.LayerList[Semester]{Key: getters.Static("semesters")}))
}
