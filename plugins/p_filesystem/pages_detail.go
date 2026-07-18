package p_filesystem

import (
	"time"

	"github.com/lariv-in/lariv"
	"github.com/lariv-in/lariv/components"
	"github.com/lariv-in/lariv/getters"
	"github.com/lariv-in/lariv/registry"
)

func pageEntriesDetail() []registry.Pair[string, components.PageInterface] {
	return []registry.Pair[string, components.PageInterface]{
		{Key: "filesystem.VNodeDetail", Value: &components.ShellScaffold{
			Sidebar: filesystemSidebar(),
			Children: []components.PageInterface{
				&components.Detail[VNode]{
					Getter: getters.Key[VNode]("vnode"),
					Children: []components.PageInterface{
						&components.ContainerColumn{Children: []components.PageInterface{
							&components.FieldTitle{Getter: getters.Key[string]("$in.Name")},
							&components.FieldSubtitle{Getter: currentVNodeTitle()},
							&components.LabelInline{Title: "Path", Classes: "mt-2", Children: []components.PageInterface{&components.FieldText{Getter: getters.Key[string]("$in.ResolvedPath")}}},
							&components.LabelInline{Title: "Type", Children: []components.PageInterface{&components.FieldText{Getter: vnodeTypeForKey("$in")}}},
							&components.LabelInline{Title: "Size", Children: []components.PageInterface{&components.FieldText{Getter: vnodeSizeForKey("$in")}}},
							&components.LabelInline{Title: "Items", Children: []components.PageInterface{&components.FieldText{Getter: getters.Key[string]("$in.ListChildrenCount")}}},
							&components.LabelInline{Title: "Created At", Children: []components.PageInterface{&components.FieldDatetime{Getter: getters.Key[time.Time]("$in.CreatedAt")}}},
							&components.LabelInline{Title: "Modified At", Children: []components.PageInterface{&components.FieldDatetime{Getter: getters.Key[time.Time]("$in.UpdatedAt")}}},
							&components.ButtonDownload{Label: "Download", Link: lariv.RoutePath("filesystem.DownloadRoute", map[string]getters.Getter[any]{
								"id": getters.Any(getters.Key[uint]("$in.ID")),
							}), Icon: "arrow-down-tray", Classes: "btn-primary mt-4"},
						}},
					},
				},
			},
		}},
	}
}
