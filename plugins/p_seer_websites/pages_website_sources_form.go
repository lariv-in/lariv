package p_seer_websites

import (
	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
)

func registerWebsiteSourceFormPages() {
	createFormName := getters.Static("seer_websites.WebsiteSourceCreateForm")
	updateFormName := getters.Static("seer_websites.WebsiteSourceUpdateForm")
	deleteFormName := getters.Static("seer_websites.WebsiteSourceDeleteForm")

	lago.RegistryPage.Register("seer_websites.WebsiteSourceCreateForm", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "seer_websites.WebsiteMenu"},
		},
		Children: []components.PageInterface{
			&components.FormListenBoostedPost{
				Name:      createFormName,
				ActionURL: lago.RoutePath("seer_websites.WebsiteSourceCreateRoute", nil),
				Children: []components.PageInterface{
					&components.FormComponent[WebsiteSource]{
						Getter:   getters.Static(WebsiteSource{}),
						Attr:     getters.FormBubbling(createFormName),
						Title:    "Create website source",
						Subtitle: "Public http(s) seed URL. Depth counts extra link hops after the seed page (same origin only). Optional worker runs this crawl on a schedule.",
						Classes:  "@container",
						ChildrenInput: []components.PageInterface{
							websiteSourceFormFields(),
						},
						ChildrenAction: []components.PageInterface{
							&components.ButtonSubmit{Label: "Save source"},
						},
					},
				},
			},
		},
	})

	lago.RegistryPage.Register("seer_websites.WebsiteSourceUpdateForm", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "seer_websites.WebsiteSourceDetailMenu"},
		},
		Children: []components.PageInterface{
			&components.FormListenBoostedPost{
				Name: updateFormName,
				ActionURL: lago.RoutePath("seer_websites.WebsiteSourceUpdateRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("websiteSource.ID")),
				}),
				Children: []components.PageInterface{
					&components.FormComponent[WebsiteSource]{
						Getter:   getters.Key[WebsiteSource]("websiteSource"),
						Attr:     getters.FormBubbling(updateFormName),
						Title:    "Edit website source",
						Subtitle: "Changing URL or depth affects the next crawl.",
						Classes:  "@container",
						ChildrenInput: []components.PageInterface{
							websiteSourceFormFields(),
						},
						ChildrenAction: []components.PageInterface{
							&components.ContainerRow{
								Classes: "flex flex-wrap justify-between gap-2 mt-2 items-center",
								Children: []components.PageInterface{
									&components.ContainerRow{
										Classes: "flex justify-end gap-2",
										Children: []components.PageInterface{
											&components.ButtonSubmit{Label: "Save changes"},
											&components.ButtonModalForm{
												Label:       "Delete",
												Icon:        "trash",
												Name:        deleteFormName,
												Url:         lago.RoutePath("seer_websites.WebsiteSourceDeleteRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("websiteSource.ID"))}),
												FormPostURL: lago.RoutePath("seer_websites.WebsiteSourceDeleteRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("websiteSource.ID"))}),
												ModalUID:    "seer-website-source-delete-modal",
												Classes:     "btn-error",
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	})

	lago.RegistryPage.Register("seer_websites.WebsiteSourceDeleteForm", &components.Modal{
		UID: "seer-website-source-delete-modal",
		Children: []components.PageInterface{
			&components.DeleteConfirmation{
				Title:   "Delete website source?",
				Message: "Stops scheduled crawls for this source; saved website rows are not removed.",
				Attr:    getters.FormBubbling(deleteFormName),
			},
		},
	})
}
