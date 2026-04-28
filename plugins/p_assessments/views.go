package p_assessments

import (
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_users"
	"github.com/lariv-in/lago/registry"
	"github.com/lariv-in/lago/views"
)

var assessmentM2MPreload = views.QueryPatchers[Assessment]{
	registry.Pair[string, views.QueryPatcher[Assessment]]{Key: "assessments.m2m", Value: views.QueryPatcherPreload[Assessment]{Fields: []string{"Topics"}}},
}

func init() {
	lago.RegistryView.Register("assessments.ListView",
		lago.GetPageView("assessments.GradeEntryTable").WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("assessments.list", views.LayerList[GradeEntry]{Key: getters.Static("grade_entries")}))
	lago.RegistryView.Register("assessments.DetailView",
		lago.GetPageView("assessments.GradeEntryDetail").WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("assessments.detail", views.LayerDetail[GradeEntry]{Key: getters.Static("grade_entry"), PathParamKey: getters.Static("id")}))
	lago.RegistryView.Register("assessments.CreateView",
		lago.GetPageView("assessments.GradeEntryCreateForm").WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("assessments.create", views.LayerCreate[GradeEntry]{SuccessURL: lago.RoutePath("assessments.DetailRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("$id"))})}))
	lago.RegistryView.Register("assessments.UpdateView",
		lago.GetPageView("assessments.GradeEntryUpdateForm").WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("assessments.detail", views.LayerDetail[GradeEntry]{Key: getters.Static("grade_entry"), PathParamKey: getters.Static("id")}).
			WithLayer("assessments.update", views.LayerUpdate[GradeEntry]{Key: getters.Static("grade_entry"), SuccessURL: lago.RoutePath("assessments.DetailRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("grade_entry.ID"))})}))
	lago.RegistryView.Register("assessments.DeleteView",
		lago.GetPageView("assessments.GradeEntryDeleteForm").WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("assessments.detail", views.LayerDetail[GradeEntry]{Key: getters.Static("grade_entry"), PathParamKey: getters.Static("id")}).
			WithLayer("assessments.delete", views.LayerDelete[GradeEntry]{Key: getters.Static("grade_entry"), SuccessURL: lago.RoutePath("assessments.DefaultRoute", nil)}))

	lago.RegistryView.Register("assessments.ExamListView",
		lago.GetPageView("assessments.AssessmentTable").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("assessments.exams.list", views.LayerList[Assessment]{Key: getters.Static("assessments")}))

	lago.RegistryView.Register("assessments.ExamDetailView",
		lago.GetPageView("assessments.AssessmentDetail").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("assessments.exams.detail", views.LayerDetail[Assessment]{
				Key:           getters.Static("assessment"),
				PathParamKey:  getters.Static("id"),
				QueryPatchers: assessmentM2MPreload,
			}))

	lago.RegistryView.Register("assessments.ExamCreateView",
		lago.GetPageView("assessments.AssessmentCreateForm").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("assessments.exams.multistep", views.MultiStepFormLayer{}).
			WithLayer("assessments.exams.create", views.LayerCreate[Assessment]{
				SuccessURL: lago.RoutePath("assessments.ExamDetailRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("$id"))}),
			}))

	lago.RegistryView.Register("assessments.ExamUpdateView",
		lago.GetPageView("assessments.AssessmentUpdateForm").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("assessments.exams.detail", views.LayerDetail[Assessment]{
				Key:           getters.Static("assessment"),
				PathParamKey:  getters.Static("id"),
				QueryPatchers: assessmentM2MPreload,
			}).
			WithLayer("assessments.exams.multistep", views.MultiStepFormLayer{}).
			WithLayer("assessments.exams.update", views.LayerUpdate[Assessment]{
				Key: getters.Static("assessment"),
				SuccessURL: lago.RoutePath("assessments.ExamDetailRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("assessment.ID")),
				}),
			}))

	lago.RegistryView.Register("assessments.ExamDeleteView",
		lago.GetPageView("assessments.AssessmentDeleteForm").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("assessments.exams.detail", views.LayerDetail[Assessment]{Key: getters.Static("assessment"), PathParamKey: getters.Static("id")}).
			WithLayer("assessments.exams.delete", views.LayerDelete[Assessment]{
				Key:        getters.Static("assessment"),
				SuccessURL: lago.RoutePath("assessments.ExamDefaultRoute", nil),
			}))
}
