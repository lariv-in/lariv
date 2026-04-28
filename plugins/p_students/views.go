package p_students

import (
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_users"
	"github.com/lariv-in/lago/registry"
	"github.com/lariv-in/lago/views"
)

var studentM2MPreload = views.QueryPatchers[Student]{
	registry.Pair[string, views.QueryPatcher[Student]]{Key: "students.m2m", Value: views.QueryPatcherPreload[Student]{Fields: []string{"Documents"}}},
}

func init() {
	lago.RegistryView.Register("students.ListView",
		lago.GetPageView("students.StudentTable").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("students.list", views.LayerList[Student]{
				Key: getters.Static("students"),
			}))

	lago.RegistryView.Register("students.DetailView",
		lago.GetPageView("students.StudentDetail").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("students.detail", views.LayerDetail[Student]{
				Key:           getters.Static("student"),
				PathParamKey:  getters.Static("id"),
				QueryPatchers: studentM2MPreload,
			}))

	lago.RegistryView.Register("students.CreateView",
		lago.GetPageView("students.StudentCreateForm").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("students.create", views.LayerCreate[Student]{
				SuccessURL: lago.RoutePath("students.DetailRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("$id")),
				}),
			}))

	lago.RegistryView.Register("students.UpdateView",
		lago.GetPageView("students.StudentUpdateForm").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("students.detail", views.LayerDetail[Student]{
				Key:           getters.Static("student"),
				PathParamKey:  getters.Static("id"),
				QueryPatchers: studentM2MPreload,
			}).
			WithLayer("students.update", views.LayerUpdate[Student]{
				Key: getters.Static("student"),
				SuccessURL: lago.RoutePath("students.DetailRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("student.ID")),
				}),
			}))

	lago.RegistryView.Register("students.DeleteView",
		lago.GetPageView("students.StudentDeleteForm").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("students.detail", views.LayerDetail[Student]{
				Key:          getters.Static("student"),
				PathParamKey: getters.Static("id"),
			}).
			WithLayer("students.delete", views.LayerDelete[Student]{
				Key:        getters.Static("student"),
				SuccessURL: lago.RoutePath("students.DefaultRoute", nil),
			}))

	lago.RegistryView.Register("students.SelectView",
		lago.GetPageView("students.StudentSelectionTable").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("students.select", views.LayerList[Student]{
				Key: getters.Static("students"),
			}))

	lago.RegistryView.Register("students.MultiSelectView",
		lago.GetPageView("students.StudentMultiSelectionTable").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("students.multiselect", views.LayerList[Student]{Key: getters.Static("students")}))
}
