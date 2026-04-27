package p_events

import (
	"time"

	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
)

func registerTablePages() {
	lago.RegistryPage.Register("events.SchoolEventTable", &components.ShellScaffold{
		Sidebar: []components.PageInterface{lago.DynamicPage{Name: "events.SchoolEventMenu"}},
		Children: []components.PageInterface{
			&components.DataTable[SchoolEvent]{
				Page:    components.Page{Key: "events.SchoolEventTableBody"},
				UID:     "school-event-table",
				Classes: "w-full",
				Data:    getters.Key[components.ObjectList[SchoolEvent]]("school_events"),
				Actions: []components.PageInterface{&components.TableButtonCreate{Link: lago.RoutePath("events.CreateRoute", nil)}},
				RowAttr: getters.RowAttrNavigate(lago.RoutePath("events.DetailRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("$row.ID"))})),
				Columns: []components.TableColumn{
					{Label: "Title", Name: "Title", Children: []components.PageInterface{&components.FieldText{Getter: getters.Key[string]("$row.Title")}}},
					{Label: "Starts", Name: "StartsAt", Children: []components.PageInterface{&components.FieldDatetime{Getter: getters.Key[time.Time]("$row.StartsAt")}}},
					{Label: "Ends", Name: "EndsAt", Children: []components.PageInterface{&components.FieldDatetime{Getter: getters.Deref(getters.Key[*time.Time]("$row.EndsAt"))}}},
				},
			},
		},
	})
}
