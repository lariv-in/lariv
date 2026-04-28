package p_finances

import (
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_users"
	"github.com/lariv-in/lago/registry"
	"github.com/lariv-in/lago/views"
)

var studentChargePreload = views.QueryPatchers[StudentCharge]{
	registry.Pair[string, views.QueryPatcher[StudentCharge]]{Key: "finances.fk", Value: views.QueryPatcherPreload[StudentCharge]{Fields: []string{"Student", "Semester"}}},
}

func init() {
	lago.RegistryView.Register("finances.ListView",
		lago.GetPageView("finances.StudentChargeTable").WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("finances.list", views.LayerList[StudentCharge]{Key: getters.Static("student_charges")}))
	lago.RegistryView.Register("finances.DetailView",
		lago.GetPageView("finances.StudentChargeDetail").WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("finances.detail", views.LayerDetail[StudentCharge]{Key: getters.Static("student_charge"), PathParamKey: getters.Static("id"), QueryPatchers: studentChargePreload}))
	lago.RegistryView.Register("finances.CreateView",
		lago.GetPageView("finances.StudentChargeCreateForm").WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("finances.create", views.LayerCreate[StudentCharge]{SuccessURL: lago.RoutePath("finances.DetailRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("$id"))})}))
	lago.RegistryView.Register("finances.UpdateView",
		lago.GetPageView("finances.StudentChargeUpdateForm").WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("finances.detail", views.LayerDetail[StudentCharge]{Key: getters.Static("student_charge"), PathParamKey: getters.Static("id"), QueryPatchers: studentChargePreload}).
			WithLayer("finances.update", views.LayerUpdate[StudentCharge]{Key: getters.Static("student_charge"), SuccessURL: lago.RoutePath("finances.DetailRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("student_charge.ID"))})}))
	lago.RegistryView.Register("finances.DeleteView",
		lago.GetPageView("finances.StudentChargeDeleteForm").WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("finances.detail", views.LayerDetail[StudentCharge]{Key: getters.Static("student_charge"), PathParamKey: getters.Static("id")}).
			WithLayer("finances.delete", views.LayerDelete[StudentCharge]{Key: getters.Static("student_charge"), SuccessURL: lago.RoutePath("finances.DefaultRoute", nil)}))
}
