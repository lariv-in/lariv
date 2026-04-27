package p_programs

import (
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_users"
	"github.com/lariv-in/lago/registry"
	"github.com/lariv-in/lago/views"
)

var programM2MPreload = views.QueryPatchers[Program]{
	registry.Pair[string, views.QueryPatcher[Program]]{Key: "programs.m2m", Value: views.QueryPatcherPreload[Program]{Fields: []string{"Students", "Teachers"}}},
}

func init() {
	lago.RegistryView.Register("programs.ListView",
		lago.GetPageView("programs.ProgramTable").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("programs.list", views.LayerList[Program]{Key: getters.Static("programs")}))

	lago.RegistryView.Register("programs.DetailView",
		lago.GetPageView("programs.ProgramDetail").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("programs.detail", views.LayerDetail[Program]{
				Key:           getters.Static("program"),
				PathParamKey:  getters.Static("id"),
				QueryPatchers: programM2MPreload,
			}))

	lago.RegistryView.Register("programs.CreateView",
		lago.GetPageView("programs.ProgramCreateForm").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("programs.create", views.LayerCreate[Program]{SuccessURL: lago.RoutePath("programs.DetailRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("$id"))})}))

	lago.RegistryView.Register("programs.UpdateView",
		lago.GetPageView("programs.ProgramUpdateForm").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("programs.detail", views.LayerDetail[Program]{
				Key:           getters.Static("program"),
				PathParamKey:  getters.Static("id"),
				QueryPatchers: programM2MPreload,
			}).
			WithLayer("programs.update", views.LayerUpdate[Program]{Key: getters.Static("program"), SuccessURL: lago.RoutePath("programs.DetailRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("program.ID"))})}))

	lago.RegistryView.Register("programs.DeleteView",
		lago.GetPageView("programs.ProgramDeleteForm").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("programs.detail", views.LayerDetail[Program]{Key: getters.Static("program"), PathParamKey: getters.Static("id")}).
			WithLayer("programs.delete", views.LayerDelete[Program]{Key: getters.Static("program"), SuccessURL: lago.RoutePath("programs.DefaultRoute", nil)}))

	lago.RegistryView.Register("programs.SelectView",
		lago.GetPageView("programs.ProgramSelectionTable").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("programs.select", views.LayerList[Program]{Key: getters.Static("programs")}))

	lago.RegistryView.Register("programs.MultiSelectView",
		lago.GetPageView("programs.ProgramMultiSelectionTable").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("programs.multiselect", views.LayerList[Program]{Key: getters.Static("programs")}))
}
