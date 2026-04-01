package p_nirmancampus_website

import (
	"context"

	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_users"
	"maragu.dev/gomponents"
)

var websiteAppRoleMiddleware = p_users.RoleAuthorizationMiddleware([]string{"admin"})

type websiteAppLandingPage struct {
	components.Page
}

func (p *websiteAppLandingPage) Build(ctx context.Context) gomponents.Node {
	return components.Render(components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "nirmancampus_website.WebsiteAdminMenu"},
		},
		Children: []components.PageInterface{
			components.ContainerColumn{
				Classes: "max-w-3xl",
				Children: []components.PageInterface{
					&components.FieldTitle{Getter: getters.Static("Website Admin")},
					&components.FieldText{Getter: getters.Static("Use the sidebar to navigate.")},
					components.ContainerRow{
						Classes: "flex gap-2 flex-wrap mt-4",
						Children: []components.PageInterface{
							&components.ButtonLink{
								Label:   "View Website",
								Link:    getters.Static("/"),
								Classes: "btn btn-outline",
							},
						},
					},
				},
			},
		},
	}, ctx)
}

func (p *websiteAppLandingPage) GetKey() string     { return p.Key }
func (p *websiteAppLandingPage) GetRoles() []string { return p.Roles }

func init() {
	lago.RegistryPage.Register("nirmancampus_website.AppLandingPage", &websiteAppLandingPage{})

	lago.RegistryPage.Register("nirmancampus_website.WebsiteAdminMenu", &components.SidebarMenu{
		Title: getters.Static("Website"),
		Back: &components.SidebarMenuItem{
			Title: getters.Static("Back to All Apps"),
			Url:   lago.GetterRoutePath("dashboard.AppsPage", nil),
		},
		Children: []components.PageInterface{
			&components.SidebarMenuItem{
				Title: getters.Static("Home"),
				Url:   lago.GetterRoutePath("nirmancampus_website.AppLandingRoute", nil),
			},
			&components.SidebarMenuItem{
				Title: getters.Static("Student Zone Sections"),
				Url:   lago.GetterRoutePath("nirmancampus_website.StudentZoneAdminDefaultRoute", nil),
			},
			&components.SidebarMenuItem{
				Title: getters.Static("Student Zone Items"),
				Url:   lago.GetterRoutePath("nirmancampus_website.StudentZoneAdminItemListRoute", nil),
			},
			&components.SidebarMenuItem{
				Title: getters.Static("Important Links"),
				Url:   lago.GetterRoutePath("nirmancampus_website.ImportantLinksDefaultRoute", nil),
			},
		},
	})

	_ = lago.RegistryRoute.Register("nirmancampus_website.AppLandingRoute", lago.Route{
		Path:    AppUrl,
		Handler: lago.NewDynamicView("nirmancampus_website.AppLandingView"),
	})

	lago.RegistryView.Register("nirmancampus_website.AppLandingView",
		lago.GetPageView("nirmancampus_website.AppLandingPage").
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware).
			WithMiddleware("website.role", websiteAppRoleMiddleware))
}
