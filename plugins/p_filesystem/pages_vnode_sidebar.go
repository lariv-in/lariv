package p_filesystem

import (
	"context"
	"fmt"

	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
)

func filesystemSidebar() []components.PageInterface {
	return []components.PageInterface{
		&components.ShowIf{
			Getter: getters.Map(getters.Key[VNode]("vnode"), func(_ context.Context, n VNode) (any, error) {
				return n.ID != 0, nil
			}),
			Children: []components.PageInterface{
				lago.DynamicPage{Name: "filesystem.VNodeMenu"},
			},
		},
		&components.ShowIf{
			Getter: getters.Map(getters.Key[VNode]("vnode"), func(_ context.Context, n VNode) (any, error) {
				return n.ID == 0, nil
			}),
			Children: []components.PageInterface{
				lago.DynamicPage{Name: "filesystem.MainMenu"},
			},
		},
	}
}

func vnodeFormFields(includeFileInput bool, submitLabel string, multi bool) components.ContainerColumn {
	initialIsDirectory := false
	isDirectoryGetter := getters.Key[bool]("$in.IsDirectory")
	if v, err := isDirectoryGetter(context.Background()); err == nil {
		initialIsDirectory = v
	}

	children := []components.PageInterface{
		&components.ContainerError{
			Error: getters.Key[error]("$error.ParentID"),
			Children: []components.PageInterface{
				&components.InputForeignKey[VNode]{
					Label:       "Location",
					Name:        "ParentID",
					Getter:      currentLocationGetter(),
					Url:         listOrBrowseRoute("filesystem.SelectRoute", "filesystem.SelectChildRoute"),
					Display:     getters.Key[string]("$in.Name"),
					Placeholder: "Root (no parent)",
				},
			},
		},
		&components.ContainerError{
			Error: getters.Key[error]("$error.IsDirectory"),
			Children: []components.PageInterface{
				&components.InputCheckbox{Label: "Create Directory", Name: "IsDirectory", Getter: getters.Key[bool]("$in.IsDirectory"), XModel: "isDirectory"},
			},
		},
		&components.ClientIf{
			Condition: "isDirectory",
			Children: []components.PageInterface{&components.ContainerError{
				Error: getters.Key[error]("$error.Name"),
				Children: []components.PageInterface{
					&components.InputText{Label: "Name", Name: "Name", Getter: getters.Key[string]("$in.Name")},
				},
			}},
		},
	}

	if includeFileInput {
		label := "File"
		name := "File"
		if multi {
			label = "Files"
			name = "Files"
		}
		children = append(children, &components.ClientIf{
			Condition: "!isDirectory",
			Children: []components.PageInterface{&components.ContainerError{
				Error: getters.Key[error]("$error." + name),
				Children: []components.PageInterface{
					&components.InputFile{Label: label, Name: name, Multiple: multi},
				},
			}},
		})
	}

	if submitLabel != "" {
		children = append(children, &components.ButtonSubmit{Label: submitLabel})
	}
	return components.ContainerColumn{
		Children: []components.PageInterface{
			&components.ClientData{
				Data:     fmt.Sprintf(`{ isDirectory: %t }`, initialIsDirectory),
				Children: children,
			},
		},
	}
}
