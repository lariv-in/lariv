package p_lacerate

import (
	"context"
	"strings"

	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
)

func registerTargetOfInterestPages() {
	registerTargetOfInterestLacerateMenuPatch()
	registerTargetOfInterestMenus()
	registerTargetOfInterestTable()
	registerTargetOfInterestForms()
	registerTargetOfInterestDetail()
}

func registerTargetOfInterestLacerateMenuPatch() {
	lago.RegistryPage.Patch("lacerate.LacerateMenu", func(page components.PageInterface) components.PageInterface {
		menu, ok := page.(*components.SidebarMenu)
		if !ok {
			return page
		}
		menu.Children = append(menu.Children, &components.SidebarMenuItem{
			Title: getters.Static("Targets of interest"),
			Url:   lago.RoutePath("lacerate.TargetOfInterestListRoute", nil),
		})
		return menu
	})
}

func registerTargetOfInterestMenus() {
	lago.RegistryPage.Register("lacerate.TargetOfInterestDetailMenu", &components.SidebarMenu{
		Title: getters.Format("Target of interest — %s", getters.Any(getters.Key[string]("target_of_interest.Name"))),
		Back: &components.SidebarMenuItem{
			Title: getters.Static("Back to targets of interest"),
			Url:   lago.RoutePath("lacerate.TargetOfInterestListRoute", nil),
		},
		Children: []components.PageInterface{
			&components.SidebarMenuItem{
				Title: getters.Static("Detail"),
				Url: lago.RoutePath("lacerate.TargetOfInterestDetailRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("target_of_interest.ID")),
				}),
			},
			&components.SidebarMenuItem{
				Title: getters.Static("Edit"),
				Url: lago.RoutePath("lacerate.TargetOfInterestUpdateRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("target_of_interest.ID")),
				}),
			},
		},
	})
}

func targetOfInterestFormFields() components.PageInterface {
	return &components.ContainerColumn{
		Page: components.Page{Key: "lacerate.TargetOfInterestFormFields"},
		Children: []components.PageInterface{
			&components.ContainerError{
				Error: getters.Key[error]("$error.Name"),
				Children: []components.PageInterface{
					&components.InputText{
						Label:    "Name",
						Name:     "Name",
						Required: true,
						Getter:   getters.Key[string]("$in.Name"),
					},
				},
			},
			&components.ContainerError{
				Error: getters.Key[error]("$error.Description"),
				Children: []components.PageInterface{
					&components.InputTextarea{
						Label:   "Description",
						Name:    "Description",
						Rows:    8,
						Classes: "w-full",
						Getter:  getters.Key[string]("$in.Description"),
					},
				},
			},
		},
	}
}

func registerTargetOfInterestTable() {
	lago.RegistryPage.Register("lacerate.TargetsOfInterestTable", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "lacerate.LacerateMenu"},
		},
		Children: []components.PageInterface{
			&components.DataTable[TargetOfInterest]{
				Page:    components.Page{Key: "lacerate.TargetsOfInterestTableBody"},
				UID:     "lacerate-targets-of-interest-table",
				Classes: "w-full",
				Data:    getters.Key[components.ObjectList[TargetOfInterest]]("targets_of_interest"),
				Actions: []components.PageInterface{
					&components.TableButtonCreate{Link: lago.RoutePath("lacerate.TargetOfInterestCreateRoute", nil)},
				},
				RowAttr: getters.RowAttrNavigate(
					lago.RoutePath("lacerate.TargetOfInterestDetailRoute", map[string]getters.Getter[any]{
						"id": getters.Any(getters.Key[uint]("$row.ID")),
					}),
				),
				Columns: []components.TableColumn{
					{
						Label: "Name",
						Children: []components.PageInterface{
							&components.FieldText{Getter: getters.Key[string]("$row.Name")},
						},
					},
					{
						Label: "Description",
						Children: []components.PageInterface{
							&components.FieldText{Getter: getters.IfOrElse(
								getters.Map(getters.Key[string]("$row.Description"), func(_ context.Context, s string) (string, error) {
									s = strings.TrimSpace(s)
									if s == "" {
										return "", nil
									}
									if len(s) > 96 {
										return s[:93] + "...", nil
									}
									return s, nil
								}),
								getters.Static("—"),
							)},
						},
					},
				},
			},
		},
	})
}

func registerTargetOfInterestForms() {
	createName := getters.Static("lacerate.TargetOfInterestCreateForm")
	updateName := getters.Static("lacerate.TargetOfInterestUpdateForm")
	deleteName := getters.Static("lacerate.TargetOfInterestDeleteForm")

	lago.RegistryPage.Register("lacerate.TargetOfInterestCreateForm", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "lacerate.LacerateMenu"},
		},
		Children: []components.PageInterface{
			&components.FormListenBoostedPost{
				Name:      createName,
				ActionURL: lago.RoutePath("lacerate.TargetOfInterestCreateRoute", nil),
				Children: []components.PageInterface{
					&components.FormComponent[TargetOfInterest]{
						Attr:     getters.FormBubbling(createName),
						Title:    "New target of interest",
						Subtitle: "Short, accurate entity summary; embedding refreshes on save when configured.",
						Classes:  "@container",
						ChildrenInput: []components.PageInterface{
							targetOfInterestFormFields(),
						},
						ChildrenAction: []components.PageInterface{
							&components.ButtonSubmit{Label: "Create"},
						},
					},
				},
			},
		},
	})

	lago.RegistryPage.Register("lacerate.TargetOfInterestUpdateForm", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "lacerate.TargetOfInterestDetailMenu"},
		},
		Children: []components.PageInterface{
			&components.FormListenBoostedPost{
				Name: updateName,
				ActionURL: lago.RoutePath("lacerate.TargetOfInterestUpdateRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("target_of_interest.ID")),
				}),
				Children: []components.PageInterface{
					&components.FormComponent[TargetOfInterest]{
						Getter:  getters.Key[TargetOfInterest]("target_of_interest"),
						Attr:    getters.FormBubbling(updateName),
						Title:   "Edit target of interest",
						Classes: "@container",
						ChildrenInput: []components.PageInterface{
							targetOfInterestFormFields(),
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
												Label: "Delete",
												Icon:  "trash",
												Name:  deleteName,
												Url: lago.RoutePath("lacerate.TargetOfInterestDeleteRoute", map[string]getters.Getter[any]{
													"id": getters.Any(getters.Key[uint]("target_of_interest.ID")),
												}),
												FormPostURL: lago.RoutePath("lacerate.TargetOfInterestDeleteRoute", map[string]getters.Getter[any]{
													"id": getters.Any(getters.Key[uint]("target_of_interest.ID")),
												}),
												ModalUID: "lacerate-target-of-interest-delete-modal",
												Classes:  "btn-error",
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
}

func registerTargetOfInterestDetail() {
	lago.RegistryPage.Register("lacerate.TargetOfInterestDetail", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "lacerate.TargetOfInterestDetailMenu"},
		},
		Children: []components.PageInterface{
			&components.Detail[TargetOfInterest]{
				Getter: getters.Key[TargetOfInterest]("target_of_interest"),
				Children: []components.PageInterface{
					components.ContainerColumn{
						Page: components.Page{Key: "lacerate.TargetOfInterestDetailContent"},
						Children: []components.PageInterface{
							&components.FieldTitle{Getter: getters.Key[string]("$in.Name")},
							&components.LabelInline{
								Title: "Description",
								Children: []components.PageInterface{
									&components.FieldMarkdown{
										Getter:  getters.Key[string]("$in.Description"),
										Classes: "prose prose-sm max-w-none",
									},
								},
							},
						},
					},
				},
			},
		},
	})

	lago.RegistryPage.Register("lacerate.TargetOfInterestDeleteForm", &components.Modal{
		UID: "lacerate-target-of-interest-delete-modal",
		Children: []components.PageInterface{
			&components.DeleteConfirmation{
				Title:   "Delete target of interest",
				Message: "Delete this target of interest? This cannot be undone.",
				Attr:    getters.FormBubbling(getters.Key[string]("$get.name")),
			},
		},
	})
}
