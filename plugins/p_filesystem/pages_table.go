package p_filesystem

import (
	"time"

	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
)

func registerTables() {
	lago.RegistryPage.Register("filesystem.VNodeTable", &components.ShellScaffold{
		Sidebar: filesystemSidebar(),
		Children: []components.PageInterface{
			&components.DataTable[VNode]{
				UID:      "filesystem-table",
				Data:     getters.Key[components.ObjectList[VNode]]("vnodes"),
				Title:    "Filesystem",
				Subtitle: "Files and folders",
				Actions: []components.PageInterface{
					&components.TableButtonFilter{Child: lago.DynamicPage{Name: "filesystem.VNodeFilter"}},
					&components.TableButtonCreate{Link: listOrBrowseRoute("filesystem.CreateRoute", "filesystem.CreateChildRoute")},
				},
				RowAttr: getters.RowAttrNavigate(rowOpenRoute()),
				Columns: []components.TableColumn{
					{Label: "Name", Name: "Name", Children: []components.PageInterface{&components.FieldText{Getter: getters.Key[string]("$row.Name")}}},
					{Label: "Type", Name: "Type", Children: []components.PageInterface{&components.FieldText{Getter: vnodeTypeForKey("$row")}}},
					{Label: "Size", Name: "Size", Children: []components.PageInterface{&components.FieldText{Getter: vnodeSizeForKey("$row")}}},
					{Label: "Items", Name: "Items", Children: []components.PageInterface{&components.FieldText{Getter: getters.Key[string]("$row.ListChildrenCount")}}},
					{Label: "Modified", Name: "UpdatedAt", Children: []components.PageInterface{&components.FieldDatetime{Getter: getters.Key[time.Time]("$row.UpdatedAt")}}},
				},
			},
		},
	})
}
