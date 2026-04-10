package p_lacerate

import (
	"context"
	"log/slog"
	"strings"

	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/registry"
)

func registerTargetOfInterestPages() {
	registerTargetOfInterestlacerateMenuPatch()
	registerTargetOfInterestMenus()
	registerTargetOfInterestTable()
	registerTargetOfInterestForms()
	registerTargetOfInterestDetail()
}

func registerTargetOfInterestlacerateMenuPatch() {
	lago.RegistryPage.Patch("lacerate.LacerateMenu", func(page components.PageInterface) components.PageInterface {
		menu, ok := page.(*components.SidebarMenu)
		if !ok {
			return page
		}
		menu.Children = append(menu.Children, &components.SidebarMenuItem{
			Title: getters.Static("Targets of Interest"),
			Url:   lago.RoutePath("lacerate.TargetOfInterestListRoute", nil),
		})
		return menu
	})
}

func registerTargetOfInterestMenus() {
	lago.RegistryPage.Register("lacerate.TargetOfInterestDetailMenu", &components.SidebarMenu{
		Title: getters.Format("Target of Interest — %s", getters.Any(getters.Key[string]("targetOfInterest.Name"))),
		Back: &components.SidebarMenuItem{
			Title: getters.Static("Back to Targets of Interest"),
			Url:   lago.RoutePath("lacerate.TargetOfInterestListRoute", nil),
		},
		Children: []components.PageInterface{
			&components.SidebarMenuItem{
				Title: getters.Static("Detail"),
				Url: lago.RoutePath("lacerate.TargetOfInterestDetailRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("targetOfInterest.ID")),
				}),
			},
			&components.SidebarMenuItem{
				Title: getters.Static("Edit"),
				Url: lago.RoutePath("lacerate.TargetOfInterestUpdateRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("targetOfInterest.ID")),
				}),
			},
		},
	})
}

func targetOfInterestTypePairGetter() getters.Getter[registry.Pair[string, string]] {
	return func(ctx context.Context) (registry.Pair[string, string], error) {
		s, err := getters.Key[string]("$in.Type")(ctx)
		if err != nil || s == "" {
			return registry.Pair[string, string]{}, nil
		}
		if p, ok := registry.PairFromPairs(s, TargetOfInterestTypeChoices); ok {
			return p, nil
		}
		return registry.Pair[string, string]{Key: s, Value: s}, nil
	}
}

func targetOfInterestContentPreviewCell() getters.Getter[string] {
	return func(ctx context.Context) (string, error) {
		s, err := getters.Key[string]("$row.Content")(ctx)
		if err != nil {
			slog.Error("lacerate: target of interest content preview cell", "error", err)
			return "", err
		}
		s = strings.TrimSpace(s)
		if s == "" {
			return "—", nil
		}
		if len(s) > 96 {
			return s[:93] + "...", nil
		}
		return s, nil
	}
}

func targetOfInterestTypeCellGetter() getters.Getter[string] {
	return func(ctx context.Context) (string, error) {
		s, err := getters.Key[string]("$row.Type")(ctx)
		if err != nil {
			slog.Error("lacerate: target of interest type cell getter", "error", err)
			return "", err
		}
		if p, ok := registry.PairFromPairs(s, TargetOfInterestTypeChoices); ok {
			return p.Value, nil
		}
		if s == "" {
			return "—", nil
		}
		return s, nil
	}
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
						Rows:    4,
						Classes: "w-full",
						Getter:  getters.Key[string]("$in.Description"),
					},
				},
			},
			&components.ContainerError{
				Error: getters.Key[error]("$error.Type"),
				Children: []components.PageInterface{
					&components.InputSelect[string]{
						Label:    "Type",
						Name:     "Type",
						Required: true,
						Choices:  getters.Static(TargetOfInterestTypeChoices),
						Getter:   targetOfInterestTypePairGetter(),
					},
				},
			},
			&components.ContainerError{
				Error: getters.Key[error]("$error.Content"),
				Children: []components.PageInterface{
					&components.InputTextarea{
						Label:   "Content",
						Name:    "Content",
						Rows:    14,
						Classes: "w-full font-mono text-sm",
						Getter:  getters.Key[string]("$in.Content"),
					},
				},
			},
		},
	}
}

func registerTargetOfInterestTable() {
	lago.RegistryPage.Register("lacerate.TargetOfInterestsTable", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "lacerate.LacerateMenu"},
		},
		Children: []components.PageInterface{
			&components.DataTable[TargetOfInterest]{
				Page:    components.Page{Key: "lacerate.TargetOfInterestsTableBody"},
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
						Label: "Type",
						Children: []components.PageInterface{
							&components.FieldText{Getter: targetOfInterestTypeCellGetter()},
						},
					},
					{
						Label: "Content",
						Children: []components.PageInterface{
							&components.FieldText{Getter: targetOfInterestContentPreviewCell()},
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
						Title:    "New Target of Interest",
						Subtitle: "Reports and other curated content; embedding is refreshed automatically on save.",
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
					"id": getters.Any(getters.Key[uint]("targetOfInterest.ID")),
				}),
				Children: []components.PageInterface{
					&components.FormComponent[TargetOfInterest]{
						Getter:  getters.Key[TargetOfInterest]("targetOfInterest"),
						Attr:    getters.FormBubbling(updateName),
						Title:   "Edit Target of Interest",
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
													"id": getters.Any(getters.Key[uint]("targetOfInterest.ID")),
												}),
												FormPostURL: lago.RoutePath("lacerate.TargetOfInterestDeleteRoute", map[string]getters.Getter[any]{
													"id": getters.Any(getters.Key[uint]("targetOfInterest.ID")),
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
				Getter: getters.Key[TargetOfInterest]("targetOfInterest"),
				Children: []components.PageInterface{
					components.ContainerColumn{
						Page: components.Page{Key: "lacerate.TargetOfInterestDetailContent"},
						Children: []components.PageInterface{
							&components.FieldTitle{Getter: getters.Key[string]("$in.Name")},
							&components.FieldSubtitle{Getter: targetOfInterestTypeCellGetterForDetail()},
							&components.LabelInline{
								Title: "Description",
								Children: []components.PageInterface{
									&components.FieldMarkdown{
										Getter:  getters.Key[string]("$in.Description"),
										Classes: "prose prose-sm max-w-none",
									},
								},
							},
							&components.LabelInline{
								Title: "Content",
								Children: []components.PageInterface{
									&components.FieldMarkdown{
										Getter:  getters.Key[string]("$in.Content"),
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
				Title:   "Delete Target of Interest",
				Message: "Delete this Target of Interest? This cannot be undone.",
				Attr:    getters.FormBubbling(getters.Key[string]("$get.name")),
			},
		},
	})
}

func targetOfInterestTypeCellGetterForDetail() getters.Getter[string] {
	return func(ctx context.Context) (string, error) {
		s, err := getters.Key[string]("$in.Type")(ctx)
		if err != nil {
			slog.Error("lacerate: Target of Interest type detail getter", "error", err)
			return "", err
		}
		if p, ok := registry.PairFromPairs(s, TargetOfInterestTypeChoices); ok {
			return p.Value, nil
		}
		return s, nil
	}
}
