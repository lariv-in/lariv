package p_seer_websites

import (
	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
)

var websiteAddFormName = getters.Static("seer_websites.WebsiteAddForm")

func registerWebsiteFormPages() {
	lago.RegistryPage.Register("seer_websites.WebsiteAddForm", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "seer_websites.WebsiteMenu"},
		},
		Children: []components.PageInterface{
			&components.FormListenBoostedPost{
				Name:      websiteAddFormName,
				ActionURL: lago.RoutePath("seer_websites.WebsiteAddRoute", nil),
				Children: []components.PageInterface{
					&components.FormComponent[Website]{
						Getter:   getters.Static(Website{}),
						Attr:     getters.FormBubbling(websiteAddFormName),
						Title:    "Add website",
						Subtitle: "Enter a public http(s) URL. Headless Chromium loads the page; readable text is stored as markdown. Requires Chromium on the server (see LAGO_seer_websites_CHROME_BIN).",
						Classes:  "@container",
						ChildrenInput: []components.PageInterface{
							&components.InputText{
								Page:     components.Page{Key: "seer_websites.WebsiteAddURLInput"},
								Label:    "Page URL",
								Name:     "URL",
								Required: true,
								Getter:   pageURLStringFromKey("$in.URL"),
								Classes:  "w-full",
							},
						},
						ChildrenAction: []components.PageInterface{
							&components.ButtonSubmit{Label: "Scrape and save"},
						},
					},
				},
			},
		},
	})
}
