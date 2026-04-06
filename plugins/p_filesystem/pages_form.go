package p_filesystem

import (
	"context"

	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
)

func registerForms() {
	lago.RegistryPage.Register("filesystem.VNodeCreateForm", &components.ShellScaffold{
		Sidebar: filesystemSidebar(),
		Children: []components.PageInterface{
			&components.FormListenBoostedPost{
				ActionURL: listOrBrowseRoute("filesystem.CreateRoute", "filesystem.CreateChildRoute"),
				Children: []components.PageInterface{
					&components.FormComponent[VNode]{
						Attr: getters.FormBubbling(nil),

						Title:    "Create Item",
						Subtitle: "Create a new file or directory",
						ChildrenInput: []components.PageInterface{
							vnodeFormFields(true, "Create", false),
						},
					},
				},
			},
		},
	})

	lago.RegistryPage.Register("filesystem.VNodeUpdateForm", &components.ShellScaffold{
		Sidebar: filesystemSidebar(),
		Children: []components.PageInterface{
			&components.FormListenBoostedPost{
				ActionURL: currentVNodeEditRoute(),
				Children: []components.PageInterface{
					&components.FormComponent[VNode]{
						Getter: getters.Key[VNode]("vnode"),
						Attr:   getters.FormBubbling(nil),

						Title:    "Edit Item",
						Subtitle: "Update file or directory details",
						ChildrenInput: []components.PageInterface{
							&components.ContainerError{
								Error: getters.Key[error]("$error.Name"),
								Children: []components.PageInterface{
									&components.InputText{Label: "Name", Name: "Name", Getter: getters.Key[string]("$in.Name"), Required: true},
								},
							},
							&components.ShowIf{
								Getter: getters.Any(func(ctx context.Context) (bool, error) {
									isDirectory, err := getters.Key[bool]("$in.IsDirectory")(ctx)
									if err != nil {
										return false, err
									}
									return !isDirectory, nil
								}),
								Children: []components.PageInterface{
									&components.ContainerError{
										Error: getters.Key[error]("$error.File"),
										Children: []components.PageInterface{
											&components.InputFile{Label: "Replace File", Name: "File"},
										},
									},
								},
							},
						},
						ChildrenAction: []components.PageInterface{
							&components.ContainerRow{
								Classes: "flex flex-wrap justify-between gap-2 mt-2 items-center",
								Children: []components.PageInterface{
									&components.ContainerRow{
										Classes: "flex justify-end gap-2",
										Children: []components.PageInterface{
											&components.ButtonSubmit{Label: "Save"},
											&components.ButtonModalForm{
												Label:       "Delete",
												Icon:        "trash",
												Url:         currentVNodeDeleteRoute(),
												FormPostURL: currentVNodeDeleteRoute(),
												ModalUID:    "filesystem-vnode-delete-modal",
												Classes:     "btn-error",
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	})

	lago.RegistryPage.Register("filesystem.VNodeMoveForm", &components.ShellScaffold{
		Sidebar: filesystemSidebar(),
		Children: []components.PageInterface{
			&components.FormListenBoostedPost{
				ActionURL: currentVNodeMoveRoute(),
				Children: []components.PageInterface{
					&components.FormComponent[VNode]{
						Getter: getters.Key[VNode]("vnode"),
						Attr:   getters.FormBubbling(nil),

						Title:    "Move Item",
						Subtitle: "Select the destination directory",
						ChildrenInput: []components.PageInterface{
							&components.ContainerError{
								Error: getters.Key[error]("$error.DestinationID"),
								Children: []components.PageInterface{
									&components.InputForeignKey[VNode]{
										Label:       "Destination Directory",
										Name:        "DestinationID",
										Getter:      parentOfCurrentVNodeGetter(),
										Url:         lago.RoutePath("filesystem.MoveSelectRoute", nil),
										Display:     getters.Key[string]("$in.Name"),
										Placeholder: "Root (move to top level)",
									},
								},
							},
						},
						ChildrenAction: []components.PageInterface{
							&components.ButtonSubmit{Label: "Move"},
						},
					},
				},
			},
		},
	})

	lago.RegistryPage.Register("filesystem.VNodeMultiUploadForm", &components.ShellScaffold{
		Sidebar: filesystemSidebar(),
		Children: []components.PageInterface{
			&components.FormListenBoostedPost{
				ActionURL: listOrBrowseRoute("filesystem.MultiUploadRoute", "filesystem.MultiUploadChildRoute"),
				Children: []components.PageInterface{
					&components.FormComponent[VNode]{
						Attr: getters.FormBubbling(nil),

						Title:    "Bulk Upload",
						Subtitle: "Upload multiple files at once",
						ChildrenInput: []components.PageInterface{
							vnodeFormFields(false, "", false),
							&components.ContainerError{
								Error: getters.Key[error]("$error.Files"),
								Children: []components.PageInterface{
									&components.InputFile{Label: "Files", Name: "Files", Multiple: true, Required: true},
								},
							},
						},
						ChildrenAction: []components.PageInterface{
							&components.ButtonSubmit{Label: "Upload Files"},
						},
					},
				},
			},
		},
	})
}
