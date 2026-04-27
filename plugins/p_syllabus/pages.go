package p_syllabus

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
	registerSelectionPages()
}

func registerMenuPages() {
	lago.RegistryPage.Register("syllabus.SyllabusTopicMenu", &components.SidebarMenu{
		Title: getters.Static("Syllabus"),
		Back: &components.SidebarMenuItem{
			Title: getters.Static("Back to All Apps"),
			Url:   lago.RoutePath("dashboard.AppsPage", nil),
		},
		Children: []components.PageInterface{
			&components.SidebarMenuItem{
				Title: getters.Static("All Topics"),
				Url:   lago.RoutePath("syllabus.DefaultRoute", nil),
			},
			&components.SidebarMenuItem{
				Title: getters.Static("Create Topic"),
				Url:   lago.RoutePath("syllabus.CreateRoute", nil),
			},
		},
	})

	lago.RegistryPage.Register("syllabus.SyllabusTopicDetailMenu", &components.SidebarMenu{
		Title: getters.Format("Topic: %s", getters.Any(getters.Key[string]("syllabus_topic.Title"))),
		Back: &components.SidebarMenuItem{
			Title: getters.Static("Back to Syllabus"),
			Url:   lago.RoutePath("syllabus.DefaultRoute", nil),
		},
		Children: []components.PageInterface{
			&components.SidebarMenuItem{
				Title: getters.Static("Detail"),
				Url: lago.RoutePath("syllabus.DetailRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("syllabus_topic.ID")),
				}),
			},
			&components.SidebarMenuItem{
				Title: getters.Static("Edit"),
				Url: lago.RoutePath("syllabus.UpdateRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("syllabus_topic.ID")),
				}),
			},
		},
	})
}
