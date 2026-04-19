package p_seer_websites

import (
	"time"

	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
)

func websiteListColumns() []components.TableColumn {
	return []components.TableColumn{
		{
			Label: "URL",
			Children: []components.PageInterface{
				&components.FieldText{
					Getter:  websiteURLStringFromRowContext(),
					Classes: "break-all max-w-prose",
				},
			},
		},
		{
			Label: "Saved",
			Children: []components.PageInterface{
				&components.FieldDatetime{Getter: getters.Key[time.Time]("$row.CreatedAt")},
			},
		},
	}
}

func registerWebsiteTablePages() {
	lago.RegistryPage.Register("seer_websites.WebsiteTable", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "seer_websites.WebsiteMenu"},
		},
		Children: []components.PageInterface{
			&components.DataTable[Website]{
				Page:    components.Page{Key: "seer_websites.WebsiteTableBody"},
				UID:     "seer-websites-table",
				Classes: "w-full",
				Data:    getters.Key[components.ObjectList[Website]]("websites"),
				Actions: []components.PageInterface{
					&components.TableButtonCreate{Link: lago.RoutePath("seer_websites.WebsiteAddRoute", nil)},
				},
				RowAttr: getters.RowAttrNavigate(
					lago.RoutePath("seer_websites.WebsiteDetailRoute", map[string]getters.Getter[any]{
						"id": getters.Any(getters.Key[uint]("$row.ID")),
					}),
				),
				Columns: websiteListColumns(),
			},
		},
	})
}
