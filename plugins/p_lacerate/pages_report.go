package p_lacerate

import (
	"context"
	"strings"

	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/registry"
)

func registerReportPages() {
	registerReportLacerateMenuPatch()
	registerReportMenus()
	registerReportTable()
	registerReportForms()
	registerReportDetail()
}

func registerReportLacerateMenuPatch() {
	lago.RegistryPage.Patch("lacerate.LacerateMenu", func(page components.PageInterface) components.PageInterface {
		menu, ok := page.(*components.SidebarMenu)
		if !ok {
			return page
		}
		menu.Children = append(menu.Children, &components.SidebarMenuItem{
			Title: getters.Static("Reports"),
			Url:   lago.RoutePath("lacerate.ReportListRoute", nil),
		})
		return menu
	})
}

func registerReportMenus() {
	lago.RegistryPage.Register("lacerate.ReportDetailMenu", &components.SidebarMenu{
		Title: getters.Format("Report — %s", getters.Any(getters.Key[string]("report.Name"))),
		Back: &components.SidebarMenuItem{
			Title: getters.Static("Back to Reports"),
			Url:   lago.RoutePath("lacerate.ReportListRoute", nil),
		},
		Children: []components.PageInterface{
			&components.SidebarMenuItem{
				Title: getters.Static("Detail"),
				Url: lago.RoutePath("lacerate.ReportDetailRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("report.ID")),
				}),
			},
			&components.SidebarMenuItem{
				Title: getters.Static("Edit"),
				Url: lago.RoutePath("lacerate.ReportUpdateRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("report.ID")),
				}),
			},
		},
	})
}

func reportFormFields() components.PageInterface {
	return &components.ContainerColumn{
		Page: components.Page{Key: "lacerate.ReportFormFields"},
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
						Choices:  getters.Static(ReportTypeChoices),
						Getter: func(ctx context.Context) (registry.Pair[string, string], error) {
							s, err := getters.Key[string]("$in.Type")(ctx)
							if err != nil || s == "" {
								return registry.Pair[string, string]{}, nil
							}
							if p, ok := registry.PairFromPairs(s, ReportTypeChoices); ok {
								return p, nil
							}
							return registry.Pair[string, string]{Key: s, Value: s}, nil
						},
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

func registerReportTable() {
	lago.RegistryPage.Register("lacerate.ReportsTable", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "lacerate.LacerateMenu"},
		},
		Children: []components.PageInterface{
			&components.DataTable[Report]{
				Page:    components.Page{Key: "lacerate.ReportsTableBody"},
				UID:     "lacerate-reports-table",
				Classes: "w-full",
				Data:    getters.Key[components.ObjectList[Report]]("reports"),
				Actions: []components.PageInterface{
					&components.TableButtonCreate{Link: lago.RoutePath("lacerate.ReportCreateRoute", nil)},
				},
				RowAttr: getters.RowAttrNavigate(
					lago.RoutePath("lacerate.ReportDetailRoute", map[string]getters.Getter[any]{
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
							&components.FieldText{Getter: getters.IfOrElse(
								getters.Map(getters.Key[string]("$row.Type"), func(_ context.Context, s string) (string, error) {
									if p, ok := registry.PairFromPairs(s, ReportTypeChoices); ok {
										return p.Value, nil
									}
									if s == "" {
										return "", nil
									}
									return s, nil
								}),
								getters.Static("—"),
							)},
						},
					},
					{
						Label: "Content",
						Children: []components.PageInterface{
							&components.FieldText{Getter: getters.IfOrElse(
								getters.Map(getters.Key[string]("$row.Content"), func(_ context.Context, s string) (string, error) {
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

func registerReportForms() {
	createName := getters.Static("lacerate.ReportCreateForm")
	updateName := getters.Static("lacerate.ReportUpdateForm")
	deleteName := getters.Static("lacerate.ReportDeleteForm")

	lago.RegistryPage.Register("lacerate.ReportCreateForm", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "lacerate.LacerateMenu"},
		},
		Children: []components.PageInterface{
			&components.FormListenBoostedPost{
				Name:      createName,
				ActionURL: lago.RoutePath("lacerate.ReportCreateRoute", nil),
				Children: []components.PageInterface{
					&components.FormComponent[Report]{
						Attr:     getters.FormBubbling(createName),
						Title:    "New report",
						Subtitle: "Curated content; embedding is refreshed automatically on save.",
						Classes:  "@container",
						ChildrenInput: []components.PageInterface{
							reportFormFields(),
						},
						ChildrenAction: []components.PageInterface{
							&components.ButtonSubmit{Label: "Create"},
						},
					},
				},
			},
		},
	})

	lago.RegistryPage.Register("lacerate.ReportUpdateForm", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "lacerate.ReportDetailMenu"},
		},
		Children: []components.PageInterface{
			&components.FormListenBoostedPost{
				Name: updateName,
				ActionURL: lago.RoutePath("lacerate.ReportUpdateRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("report.ID")),
				}),
				Children: []components.PageInterface{
					&components.FormComponent[Report]{
						Getter:  getters.Key[Report]("report"),
						Attr:    getters.FormBubbling(updateName),
						Title:   "Edit report",
						Classes: "@container",
						ChildrenInput: []components.PageInterface{
							reportFormFields(),
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
												Url: lago.RoutePath("lacerate.ReportDeleteRoute", map[string]getters.Getter[any]{
													"id": getters.Any(getters.Key[uint]("report.ID")),
												}),
												FormPostURL: lago.RoutePath("lacerate.ReportDeleteRoute", map[string]getters.Getter[any]{
													"id": getters.Any(getters.Key[uint]("report.ID")),
												}),
												ModalUID: "lacerate-report-delete-modal",
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

func registerReportDetail() {
	lago.RegistryPage.Register("lacerate.ReportDetail", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "lacerate.ReportDetailMenu"},
		},
		Children: []components.PageInterface{
			&components.Detail[Report]{
				Getter: getters.Key[Report]("report"),
				Children: []components.PageInterface{
					components.ContainerColumn{
						Page: components.Page{Key: "lacerate.ReportDetailContent"},
						Children: []components.PageInterface{
							&components.FieldTitle{Getter: getters.Key[string]("$in.Name")},
							&components.FieldSubtitle{Getter: getters.Map(getters.Key[string]("$in.Type"), func(_ context.Context, s string) (string, error) {
								if p, ok := registry.PairFromPairs(s, ReportTypeChoices); ok {
									return p.Value, nil
								}
								return s, nil
							})},
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

	lago.RegistryPage.Register("lacerate.ReportDeleteForm", &components.Modal{
		UID: "lacerate-report-delete-modal",
		Children: []components.PageInterface{
			&components.DeleteConfirmation{
				Title:   "Delete report",
				Message: "Delete this report? This cannot be undone.",
				Attr:    getters.FormBubbling(getters.Key[string]("$get.name")),
			},
		},
	})
}
