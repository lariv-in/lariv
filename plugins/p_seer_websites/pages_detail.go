package p_seer_websites

import (
	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
)

func websiteDetailContentColumn() components.PageInterface {
	return components.ContainerColumn{
		Page: components.Page{Key: "seer_websites.WebsiteDetailContent"},
		Children: []components.PageInterface{
			&components.ContainerRow{
				Page:    components.Page{Key: "seer_websites.WebsiteDetailHeader"},
				Classes: "flex flex-wrap justify-between items-start gap-2 w-full",
				Children: []components.PageInterface{
					&components.FieldTitle{Getter: pageURLStringFromKey("$in.URL")},
					&components.ShowIf{
						Page:   components.Page{Key: "seer_websites.WebsiteDetailAddIntelWrap"},
						Getter: getters.Any(getters.Key[bool]("websiteIntelAddVisible")),
						Children: []components.PageInterface{
							&components.ButtonPost{
								Page:    components.Page{Key: "seer_websites.WebsiteDetailAddIntelBtn"},
								Label:   "Add to Intel",
								URL:     lago.RoutePath("seer_websites.WebsiteAddIntelRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("$in.ID"))}),
								Icon:    "document-plus",
								Classes: "btn-outline btn-primary btn-sm shrink-0",
							},
						},
					},
					&components.ButtonModalForm{
						Page:        components.Page{Key: "seer_websites.WebsiteDetailDeleteBtn"},
						Label:       "Delete",
						Icon:        "trash",
						Name:        getters.Static("seer_websites.WebsiteDeleteForm"),
						Url:         lago.RoutePath("seer_websites.WebsiteDeleteRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("website.ID"))}),
						FormPostURL: lago.RoutePath("seer_websites.WebsiteDeleteRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("website.ID"))}),
						ModalUID:    "seer-website-delete-modal",
						Classes:     "btn-outline btn-error btn-sm shrink-0",
					},
				},
			},
			&components.LabelInline{
				Title: "Markdown",
				Children: []components.PageInterface{
					&components.FieldMarkdown{
						Getter:  getters.Key[string]("$in.Markdown"),
						Classes: "prose prose-sm max-w-none max-h-96 overflow-y-auto",
					},
				},
			},
			&components.ShowIf{
				Page:   components.Page{Key: "seer_websites.WebsiteDetailIntelLinkWrap"},
				Getter: getters.Any(getters.Key[bool]("websiteIntelLinkVisible")),
				Children: []components.PageInterface{
					&components.LabelInline{
						Title: "Intel",
						Children: []components.PageInterface{
							&components.FieldLink{
								Page:    components.Page{Key: "seer_websites.WebsiteDetailIntelLink"},
								Href:    getters.Key[string]("websiteIntelDetailHref"),
								Label:   getters.Static("View intel"),
								Classes: "link link-primary",
							},
						},
					},
				},
			},
		},
	}
}

func registerWebsiteDetailPages() {
	lago.RegistryPage.Register("seer_websites.WebsiteDetail", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "seer_websites.WebsiteDetailMenu"},
		},
		Children: []components.PageInterface{
			&components.Detail[Website]{
				Getter: getters.Key[Website]("website"),
				Children: []components.PageInterface{
					websiteDetailContentColumn(),
				},
			},
		},
	})

	lago.RegistryPage.Register("seer_websites.WebsiteDeleteForm", &components.Modal{
		UID: "seer-website-delete-modal",
		Children: []components.PageInterface{
			&components.DeleteConfirmation{
				Title:   "Delete saved website?",
				Message: "Clears markdown in this app. URL kept for audit trail. Row is soft-deleted and hidden from lists.",
				Attr:    getters.FormBubbling(getters.Static("seer_websites.WebsiteDeleteForm")),
			},
		},
	})
}
