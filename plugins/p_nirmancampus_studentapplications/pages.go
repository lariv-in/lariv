package p_nirmancampus_studentapplications

import (
	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
)

func init() {
	registerMenuPages()
	registerFilterPages()
	registerFormPages()
	registerTablePages()
	registerDetailPages()
}

func registerMenuPages() {
	lago.RegistryPage.Register("studentapplications.ApplicationMenu", &components.SidebarMenu{
		Title: getters.Static("Student applications"),
		Back: &components.SidebarMenuItem{
			Title: getters.Static("Back to All Apps"),
			Url:   lago.RoutePath("dashboard.AppsPage", nil),
		},
		Children: []components.PageInterface{
			&components.SidebarMenuItem{
				Title: getters.Static("All applications"),
				Url:   lago.RoutePath("studentapplications.DefaultRoute", nil),
			},
		},
	})

	lago.RegistryPage.Register("studentapplications.ApplicationDetailMenu", &components.SidebarMenu{
		Title: getters.Format(
			"Application: %s",
			getters.Any(getters.IfOrElse(
				getters.Key[string]("studentapplication.StudentName"),
				getters.IfOrElse(
					getters.Key[string]("$in.StudentName"),
					getters.Static("Application"),
				),
			)),
		),
		Back: &components.SidebarMenuItem{
			Title: getters.Static("Back to all applications"),
			Url:   lago.RoutePath("studentapplications.DefaultRoute", nil),
		},
		Children: []components.PageInterface{
			&components.SidebarMenuItem{
				Title: getters.Static("Application detail"),
				Url: lago.RoutePath("studentapplications.DetailRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.IfOrElse(
						getters.Key[uint]("studentapplication.ID"),
						getters.ParseUint(getters.Key[string]("$path.id")),
					)),
				}),
			},
			&components.SidebarMenuItem{
				Page:  components.Page{Roles: []string{"admin", "superuser"}},
				Title: getters.Static("Edit application"),
				Url: lago.RoutePath("studentapplications.UpdateRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.IfOrElse(
						getters.Key[uint]("studentapplication.ID"),
						getters.ParseUint(getters.Key[string]("$path.id")),
					)),
				}),
			},
		},
	})
}
