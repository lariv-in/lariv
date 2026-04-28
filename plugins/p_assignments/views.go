package p_assignments

import (
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_users"
	"github.com/lariv-in/lago/registry"
	"github.com/lariv-in/lago/views"
)

var assignmentPreload = views.QueryPatchers[Assignment]{
	registry.Pair[string, views.QueryPatcher[Assignment]]{Key: "assignments.fk", Value: views.QueryPatcherPreload[Assignment]{Fields: []string{"Course", "Semester"}}},
}

func init() {
	lago.RegistryView.Register("assignments.ListView",
		lago.GetPageView("assignments.AssignmentTable").WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("assignments.list", views.LayerList[Assignment]{Key: getters.Static("assignments")}))
	lago.RegistryView.Register("assignments.DetailView",
		lago.GetPageView("assignments.AssignmentDetail").WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("assignments.detail", views.LayerDetail[Assignment]{Key: getters.Static("assignment"), PathParamKey: getters.Static("id"), QueryPatchers: assignmentPreload}))
	lago.RegistryView.Register("assignments.CreateView",
		lago.GetPageView("assignments.AssignmentCreateForm").WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("assignments.create", views.LayerCreate[Assignment]{SuccessURL: lago.RoutePath("assignments.DetailRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("$id"))})}))
	lago.RegistryView.Register("assignments.UpdateView",
		lago.GetPageView("assignments.AssignmentUpdateForm").WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("assignments.detail", views.LayerDetail[Assignment]{Key: getters.Static("assignment"), PathParamKey: getters.Static("id"), QueryPatchers: assignmentPreload}).
			WithLayer("assignments.update", views.LayerUpdate[Assignment]{Key: getters.Static("assignment"), SuccessURL: lago.RoutePath("assignments.DetailRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("assignment.ID"))})}))
	lago.RegistryView.Register("assignments.DeleteView",
		lago.GetPageView("assignments.AssignmentDeleteForm").WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("assignments.detail", views.LayerDetail[Assignment]{Key: getters.Static("assignment"), PathParamKey: getters.Static("id"), QueryPatchers: assignmentPreload}).
			WithLayer("assignments.delete", views.LayerDelete[Assignment]{Key: getters.Static("assignment"), SuccessURL: lago.RoutePath("assignments.DefaultRoute", nil)}))
}
