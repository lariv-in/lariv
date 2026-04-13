package p_nirmancampus_website

import (
	"context"

	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	"maragu.dev/gomponents"
)

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
