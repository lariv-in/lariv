package p_syllabus

import (
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_users"
	"github.com/lariv-in/lago/views"
)

func init() {
	lago.RegistryView.Register("syllabus.ListView",
		lago.GetPageView("syllabus.SyllabusTopicTable").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("syllabus.list", views.LayerList[SyllabusTopic]{Key: getters.Static("syllabus_topics")}))

	lago.RegistryView.Register("syllabus.DetailView",
		lago.GetPageView("syllabus.SyllabusTopicDetail").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("syllabus.detail", views.LayerDetail[SyllabusTopic]{
				Key:          getters.Static("syllabus_topic"),
				PathParamKey: getters.Static("id"),
			}))

	lago.RegistryView.Register("syllabus.CreateView",
		lago.GetPageView("syllabus.SyllabusTopicCreateForm").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("syllabus.create", views.LayerCreate[SyllabusTopic]{
				SuccessURL: lago.RoutePath("syllabus.DetailRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("$id")),
				}),
			}))

	lago.RegistryView.Register("syllabus.UpdateView",
		lago.GetPageView("syllabus.SyllabusTopicUpdateForm").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("syllabus.detail", views.LayerDetail[SyllabusTopic]{
				Key:          getters.Static("syllabus_topic"),
				PathParamKey: getters.Static("id"),
			}).
			WithLayer("syllabus.update", views.LayerUpdate[SyllabusTopic]{
				Key: getters.Static("syllabus_topic"),
				SuccessURL: lago.RoutePath("syllabus.DetailRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("syllabus_topic.ID")),
				}),
			}))

	lago.RegistryView.Register("syllabus.DeleteView",
		lago.GetPageView("syllabus.SyllabusTopicDeleteForm").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("syllabus.detail", views.LayerDetail[SyllabusTopic]{
				Key:          getters.Static("syllabus_topic"),
				PathParamKey: getters.Static("id"),
			}).
			WithLayer("syllabus.delete", views.LayerDelete[SyllabusTopic]{
				Key:        getters.Static("syllabus_topic"),
				SuccessURL: lago.RoutePath("syllabus.DefaultRoute", nil),
			}))

	lago.RegistryView.Register("syllabus.MultiSelectView",
		lago.GetPageView("syllabus.SyllabusTopicMultiSelectionTable").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("syllabus.multiselect", views.LayerList[SyllabusTopic]{Key: getters.Static("syllabus_topics")}))
}
