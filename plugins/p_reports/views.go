package p_reports

import (
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_users"
	"github.com/lariv-in/lago/views"
)

func init() {
	lago.RegistryView.Register("reports.ListView",
		lago.GetPageView("reports.ReportDefinitionTable").WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("reports.list", views.LayerList[ReportDefinition]{Key: getters.Static("report_definitions")}))
	lago.RegistryView.Register("reports.DetailView",
		lago.GetPageView("reports.ReportDefinitionDetail").WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("reports.detail", views.LayerDetail[ReportDefinition]{Key: getters.Static("report_definition"), PathParamKey: getters.Static("id")}))
	lago.RegistryView.Register("reports.CreateView",
		lago.GetPageView("reports.ReportDefinitionCreateForm").WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("reports.create", views.LayerCreate[ReportDefinition]{SuccessURL: lago.RoutePath("reports.DetailRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("$id"))})}))
	lago.RegistryView.Register("reports.UpdateView",
		lago.GetPageView("reports.ReportDefinitionUpdateForm").WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("reports.detail", views.LayerDetail[ReportDefinition]{Key: getters.Static("report_definition"), PathParamKey: getters.Static("id")}).
			WithLayer("reports.update", views.LayerUpdate[ReportDefinition]{Key: getters.Static("report_definition"), SuccessURL: lago.RoutePath("reports.DetailRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("report_definition.ID"))})}))
	lago.RegistryView.Register("reports.DeleteView",
		lago.GetPageView("reports.ReportDefinitionDeleteForm").WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("reports.detail", views.LayerDetail[ReportDefinition]{Key: getters.Static("report_definition"), PathParamKey: getters.Static("id")}).
			WithLayer("reports.delete", views.LayerDelete[ReportDefinition]{Key: getters.Static("report_definition"), SuccessURL: lago.RoutePath("reports.DefaultRoute", nil)}))
}
