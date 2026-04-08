package p_nirmancampus_academicrecords

import (
	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
)

func init() {
	registerMenuPages()
	registerStudentsMenuAcademicRecordsEntry()
	registerFilterPages()
	registerFormPages()
	registerTablePages()
	registerDetailPages()
	registerSelectionPages()
}

var courseDetailLink = lago.RoutePath("courses.DetailRoute", map[string]getters.Getter[any]{
	"id": getters.Any(getters.Key[uint]("$in.ID")),
})

func registerStudentsMenuAcademicRecordsEntry() {
	lago.RegistryPage.Patch("students.StudentMenu", func(page components.PageInterface) components.PageInterface {
		menu, ok := page.(*components.SidebarMenu)
		if !ok {
			return page
		}
		menu.Children = append(menu.Children, &components.SidebarMenuItem{
			Title: getters.Static("Academic Records"),
			Url:   lago.RoutePath("academicrecords.DefaultRoute", nil),
		})
		return menu
	})
}

func registerMenuPages() {
	lago.RegistryPage.Register("academicrecords.AcademicRecordDetailMenu", &components.SidebarMenu{
		Title: getters.Format("Record: %s", getters.Any(getters.Key[string]("academicrecord.Student.Name"))),
		Back: &components.SidebarMenuItem{
			Title: getters.Static("Back to all Academic Records"),
			Url:   lago.RoutePath("academicrecords.DefaultRoute", nil),
		},
		Children: []components.PageInterface{
			&components.SidebarMenuItem{
				Title: getters.Static("Academic Record Detail"),
				Url: lago.RoutePath("academicrecords.DetailRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("academicrecord.ID")),
				}),
			},
			&components.SidebarMenuItem{
				Page:  components.Page{Roles: []string{"admin", "superuser"}},
				Title: getters.Static("Edit Academic Record"),
				Url: lago.RoutePath("academicrecords.UpdateRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("academicrecord.ID")),
				}),
			},
		},
	})
}

// --- Filters ---
