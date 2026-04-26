package p_seer_deepsearch

import (
	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
)

var deepSearchHomeFormName = getters.Static("seer_deepsearch.HomeSearchForm")

func registerDeepSearchMenuPages() {
	lago.RegistryPage.Register("seer_deepsearch.DeepSearchMenu", &components.SidebarMenu{
		Title: getters.Static("Deep search"),
		Back: &components.SidebarMenuItem{
			Title: getters.Static("Back to All Apps"),
			Url:   lago.RoutePath("dashboard.AppsPage", nil),
		},
		Children: []components.PageInterface{
			&components.SidebarMenuItem{
				Title: getters.Static("New search"),
				Url:   lago.RoutePath("seer_deepsearch.DefaultRoute", nil),
			},
			&components.SidebarMenuItem{
				Title: getters.Static("History"),
				Url:   lago.RoutePath("seer_deepsearch.HistoryRoute", nil),
			},
		},
	})
}

func registerDeepSearchSearchPages() {
	lago.RegistryPage.Register("seer_deepsearch.DeepSearchHome", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "seer_deepsearch.DeepSearchMenu"},
		},
		Children: []components.PageInterface{
			&components.FormListenBoostedPost{
				Name:      deepSearchHomeFormName,
				ActionURL: lago.RoutePath("seer_deepsearch.StartRoute", nil),
				Children: []components.PageInterface{
					&components.FormComponent[DeepSearch]{
						Getter:   getters.Static(DeepSearch{}),
						Title:    "Deep search",
						Subtitle: "Enter a research question. The app expands queries, searches the web (Google Programmable Search), scrapes pages, ingests Intel, then writes a markdown report. Requires [Plugins.p_seer_deepsearch] apiKey+cx and [Plugins.p_google_genai] for LLM calls.",
						Classes:  "@container max-w-2xl mx-auto",
						ChildrenInput: []components.PageInterface{
							&components.InputText{
								Page:     components.Page{Key: "seer_deepsearch.HomeQueryInput"},
								Label:    "Question",
								Name:     "Query",
								Required: true,
								Getter:   getters.Key[string]("$in.Query"),
								Classes:  "w-full",
							},
						},
						ChildrenAction: []components.PageInterface{
							&components.ButtonSubmit{Label: "Run deep search"},
						},
						Attr: deepSearchHomeFormAttr(),
					},
				},
			},
		},
	})

	lago.RegistryPage.Register("seer_deepsearch.StartBlank", &components.ContainerColumn{
		Page: components.Page{Key: "seer_deepsearch.StartBlankRoot"},
	})
}
