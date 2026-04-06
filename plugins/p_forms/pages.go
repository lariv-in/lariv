package forms

import (
	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
)

func init() {
	registerMenus()
	registerFormListPages()
	registerFormCRUDPages()
	registerFieldPages()
	registerSubmissionPages()
	registerPublicPage()
}

func registerMenus() {
	lago.RegistryPage.Register("forms.FormMenu", &components.SidebarMenu{
		Title: getters.Static("Forms"),
		Back: &components.SidebarMenuItem{
			Title: getters.Static("Back to All Apps"),
			Url:   lago.RoutePath("dashboard.AppsPage", nil),
		},
		Children: []components.PageInterface{
			&components.SidebarMenuItem{
				Title: getters.Static("All forms"),
				Url:   lago.RoutePath("forms.DefaultRoute", nil),
			},
		},
	})

	lago.RegistryPage.Register("forms.FormDetailMenu", &components.SidebarMenu{
		Title: getters.Format("Form: %s", getters.Any(getters.Key[string]("form.Title"))),
		Back: &components.SidebarMenuItem{
			Title: getters.Static("Back to all forms"),
			Url:   lago.RoutePath("forms.DefaultRoute", nil),
		},
		Children: []components.PageInterface{
			&components.SidebarMenuItem{
				Title: getters.Static("Detail"),
				Url: lago.RoutePath("forms.DetailRoute", map[string]getters.Getter[any]{
					"form_id": getters.Any(getters.Key[uint]("form.ID")),
				}),
			},
			&components.SidebarMenuItem{
				Title: getters.Static("Edit"),
				Url: lago.RoutePath("forms.UpdateRoute", map[string]getters.Getter[any]{
					"form_id": getters.Any(getters.Key[uint]("form.ID")),
				}),
			},
			&components.SidebarMenuItem{
				Title: getters.Static("Submissions"),
				Url: lago.RoutePath("forms.SubmissionsListRoute", map[string]getters.Getter[any]{
					"form_id": getters.Any(getters.Key[uint]("form.ID")),
				}),
			},
		},
	})

	lago.RegistryPage.Register("forms.FieldEditMenu", &components.SidebarMenu{
		Title: getters.Static("Field"),
		Back: &components.SidebarMenuItem{
			Title: getters.Static("Back to form"),
			Url: lago.RoutePath("forms.DetailRoute", map[string]getters.Getter[any]{
				"form_id": getters.Any(getters.Key[uint]("form_field.FormID")),
			}),
		},
		Children: []components.PageInterface{
			&components.SidebarMenuItem{
				Title: getters.Static("Edit field"),
				Url: lago.RoutePath("forms.FieldUpdateRoute", map[string]getters.Getter[any]{
					"form_id": getters.Any(getters.Key[uint]("form_field.FormID")),
					"id":      getters.Any(getters.Key[uint]("form_field.ID")),
				}),
			},
		},
	})
}
