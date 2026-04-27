package p_assessments

import (
	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
)

func init() {
	registerMenuPages()
	registerFormPages()
	registerTablePages()
	registerDetailPages()
	registerExamPages()
}

func registerMenuPages() {
	lago.RegistryPage.Register("assessments.GradeEntryMenu", &components.SidebarMenu{
		Title: getters.Static("Assessments"),
		Back:  &components.SidebarMenuItem{Title: getters.Static("Back to All Apps"), Url: lago.RoutePath("dashboard.AppsPage", nil)},
		Children: []components.PageInterface{
			&components.SidebarMenuItem{Title: getters.Static("Exam definitions"), Url: lago.RoutePath("assessments.ExamDefaultRoute", nil)},
			&components.SidebarMenuItem{Title: getters.Static("Create exam"), Url: lago.RoutePath("assessments.ExamCreateRoute", nil)},
			&components.SidebarMenuItem{Title: getters.Static("All grades"), Url: lago.RoutePath("assessments.DefaultRoute", nil)},
			&components.SidebarMenuItem{Title: getters.Static("Create grade entry"), Url: lago.RoutePath("assessments.CreateRoute", nil)},
		},
	})
	lago.RegistryPage.Register("assessments.GradeEntryDetailMenu", &components.SidebarMenu{
		Title: getters.Format("Grade: %s", getters.Any(getters.Key[string]("grade_entry.Component"))),
		Back:  &components.SidebarMenuItem{Title: getters.Static("Back"), Url: lago.RoutePath("assessments.DefaultRoute", nil)},
		Children: []components.PageInterface{
			&components.SidebarMenuItem{Title: getters.Static("Detail"), Url: lago.RoutePath("assessments.DetailRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("grade_entry.ID"))})},
			&components.SidebarMenuItem{Title: getters.Static("Edit"), Url: lago.RoutePath("assessments.UpdateRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("grade_entry.ID"))})},
		},
	})
}
