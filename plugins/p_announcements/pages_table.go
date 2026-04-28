package p_announcements

import (
	"time"

	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
)

func registerTablePages() {
	lago.RegistryPage.Register("announcements.AnnouncementTable", &components.ShellScaffold{
		Sidebar: []components.PageInterface{lago.DynamicPage{Name: "announcements.AnnouncementMenu"}},
		Children: []components.PageInterface{
			&components.DataTable[Announcement]{
				Page: components.Page{Key: "announcements.AnnouncementTableBody"}, UID: "announcement-table", Classes: "w-full",
				Data:    getters.Key[components.ObjectList[Announcement]]("announcements"),
				Actions: []components.PageInterface{&components.TableButtonCreate{Link: lago.RoutePath("announcements.CreateRoute", nil)}},
				RowAttr: getters.RowAttrNavigate(lago.RoutePath("announcements.DetailRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("$row.ID"))})),
				Columns: []components.TableColumn{
					{Label: "Title", Name: "Title", Children: []components.PageInterface{&components.FieldText{Getter: getters.Key[string]("$row.Title")}}},
					{Label: "Release", Name: "ReleaseAt", Children: []components.PageInterface{&components.FieldDatetime{Getter: getters.Key[time.Time]("$row.ReleaseAt")}}},
					{Label: "Priority", Name: "Priority", Children: []components.PageInterface{&components.FieldText{Getter: getters.Key[string]("$row.Priority")}}},
				},
			},
		},
	})
}
