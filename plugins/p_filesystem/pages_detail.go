package p_filesystem

import (
	"context"
	"time"

	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
)

func registerDetail() {
	lago.RegistryPage.Register("filesystem.VNodeDetail", &components.ShellScaffold{
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
						&components.ShowIf{
							Getter: getters.Any(func(ctx context.Context) (bool, error) {
								isDirectory, err := getters.Key[bool]("$in.IsDirectory")(ctx)
								if err != nil {
									return false, err
								}
								return !isDirectory, nil
							}),
							Children: []components.PageInterface{
								&components.ButtonDownload{Label: "Download", Link: lago.RoutePath("filesystem.DownloadRoute", map[string]getters.Getter[any]{
									"id": getters.Any(getters.Key[uint]("$in.ID")),
								}), Icon: "arrow-down-tray", Classes: "btn-primary mt-4"},
							},
						},
					}},
				},
			},
		},
	})
}
