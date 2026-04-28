package p_forums

import (
	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
)

func registerTablePages() {
	lago.RegistryPage.Register("forums.ForumThreadTable", &components.ShellScaffold{
		Sidebar: []components.PageInterface{lago.DynamicPage{Name: "forums.ForumThreadMenu"}},
		Children: []components.PageInterface{
			&components.DataTable[ForumThread]{
				Page: components.Page{Key: "forums.ForumThreadTableBody"}, UID: "forum-thread-table", Classes: "w-full",
				Data:    getters.Key[components.ObjectList[ForumThread]]("forum_threads"),
				Actions: []components.PageInterface{&components.TableButtonCreate{Link: lago.RoutePath("forums.CreateRoute", nil)}},
				RowAttr: getters.RowAttrNavigate(lago.RoutePath("forums.DetailRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("$row.ID"))})),
				Columns: []components.TableColumn{
					{Label: "Title", Name: "Title", Children: []components.PageInterface{&components.FieldText{Getter: getters.Key[string]("$row.Title")}}},
					{Label: "Course ID", Name: "CourseID", Children: []components.PageInterface{&components.FieldText{Getter: getters.Format("%d", getters.Any(getters.Key[uint]("$row.CourseID")))}}},
				},
			},
		},
	})
}
