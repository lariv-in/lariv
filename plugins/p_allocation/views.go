package p_allocation

import (
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_users"
	"github.com/lariv-in/lago/views"
)

func init() {
	lago.RegistryView.Register("allocation.ListView",
		lago.GetPageView("allocation.CourseTeacherAssignmentTable").WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("allocation.list", views.LayerList[CourseTeacherAssignment]{Key: getters.Static("allocations")}))
	lago.RegistryView.Register("allocation.DetailView",
		lago.GetPageView("allocation.CourseTeacherAssignmentDetail").WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("allocation.detail", views.LayerDetail[CourseTeacherAssignment]{Key: getters.Static("allocation"), PathParamKey: getters.Static("id")}))
	lago.RegistryView.Register("allocation.CreateView",
		lago.GetPageView("allocation.CourseTeacherAssignmentCreateForm").WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("allocation.create", views.LayerCreate[CourseTeacherAssignment]{SuccessURL: lago.RoutePath("allocation.DetailRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("$id"))})}))
	lago.RegistryView.Register("allocation.UpdateView",
		lago.GetPageView("allocation.CourseTeacherAssignmentUpdateForm").WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("allocation.detail", views.LayerDetail[CourseTeacherAssignment]{Key: getters.Static("allocation"), PathParamKey: getters.Static("id")}).
			WithLayer("allocation.update", views.LayerUpdate[CourseTeacherAssignment]{Key: getters.Static("allocation"), SuccessURL: lago.RoutePath("allocation.DetailRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("allocation.ID"))})}))
	lago.RegistryView.Register("allocation.DeleteView",
		lago.GetPageView("allocation.CourseTeacherAssignmentDeleteForm").WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("allocation.detail", views.LayerDetail[CourseTeacherAssignment]{Key: getters.Static("allocation"), PathParamKey: getters.Static("id")}).
			WithLayer("allocation.delete", views.LayerDelete[CourseTeacherAssignment]{Key: getters.Static("allocation"), SuccessURL: lago.RoutePath("allocation.DefaultRoute", nil)}))
}
