package p_lacerate

import (
	"context"
	"strings"
	"time"

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
							&components.FieldTitle{
								Getter:  getters.Static("Locations"),
								Classes: "mt-8",
							},
							&components.DataTable[TargetOfInterestLocation]{
								Page:     components.Page{Key: "lacerate.TargetOfInterestDetailLocationsTable"},
								UID:      "lacerate-target-of-interest-locations-table",
								Subtitle: "Addresses and times for this target; each row is tied to intel (coordinates are stored but not listed here).",
								Classes:  "w-full",
								Data:     getters.Key[components.ObjectList[TargetOfInterestLocation]](ctxKeyTargetOfInterestLocations),
								RowAttr: getters.RowAttrNavigate(
									lago.RoutePath("lacerate.IntelDetailRoute", map[string]getters.Getter[any]{
										"id": getters.Any(getters.Key[uint]("$row.IntelID")),
									}),
								),
								Columns: []components.TableColumn{
									{
										Label: "Datetime",
										Children: []components.PageInterface{
											&components.FieldDatetime{Getter: getters.Key[time.Time]("$row.Datetime")},
										},
									},
									{
										Label: "Intel",
										Children: []components.PageInterface{
											&components.FieldText{
												Getter: getters.IfOrElse(
													getters.Map(getters.Key[string]("$row.Intel.Content"), func(_ context.Context, s string) (string, error) {
														s = strings.TrimSpace(s)
														if s == "" {
															return "", nil
														}
														if len(s) > 96 {
															return s[:93] + "...", nil
														}
														return s, nil
													}),
													getters.Format("Intel #%d", getters.Any(getters.Key[uint]("$row.IntelID"))),
												),
												Classes: "text-sm text-base-content/70",
											},
										},
									},
									{
										Label: "Address",
										Children: []components.PageInterface{
											&components.FieldText{Getter: getters.Key[string]("$row.Address")},
										},
									},
								},
							},
							&components.FieldTitle{
								Getter:  getters.Static("Related data"),
								Classes: "mt-8",
							},
							&components.ClientTabs{
								Page:     components.Page{Key: "lacerate.TargetOfInterestDetailRelatedTabs"},
								StateKey: "related_tab",
								Default:  getters.Static("Targets"),
								Tabs: map[string]getters.Getter[components.PageInterface]{
									"Targets": getters.Static(targetOfInterestRelatedSection()),
									"Reports": getters.Static(targetOfInterestRelatedReportsSection()),
									"Intel":   getters.Static(targetOfInterestRelatedIntelSection()),
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

func targetOfInterestRelatedSection() components.PageInterface {
	return &components.ContainerColumn{
		Page: components.Page{Key: "lacerate.TargetOfInterestDetailRelated"},
		Children: []components.PageInterface{
			&components.DataTable[TargetOfInterest]{
				Page:     components.Page{Key: "lacerate.TargetOfInterestDetailRelatedTable"},
				UID:      "lacerate-target-of-interest-related-table",
				Title:    "Related targets of interest",
				Subtitle: "Nearest embedding matches for this target.",
				Classes:  "w-full",
				Data:     getters.Key[components.ObjectList[TargetOfInterest]](ctxKeyRelatedTargetsOfInterest),
				RowAttr: getters.RowAttrNavigate(
					lago.RoutePath("lacerate.TargetOfInterestDetailRoute", map[string]getters.Getter[any]{
						"id": getters.Any(getters.Key[uint]("$row.ID")),
					}),
				),
				Columns: []components.TableColumn{
					{
						Label: "Name",
						Children: []components.PageInterface{
							&components.FieldText{
								Getter: getters.IfOrElse(
									getters.Map(getters.Key[string]("$row.Name"), func(_ context.Context, s string) (string, error) {
										return strings.TrimSpace(s), nil
									}),
									getters.Format("#%d", getters.Any(getters.Key[uint]("$row.ID"))),
								),
							},
						},
					},
					{
						Label: "Description",
						Children: []components.PageInterface{
							&components.FieldText{
								Getter: getters.IfOrElse(
									getters.Map(getters.Key[string]("$row.Description"), func(_ context.Context, s string) (string, error) {
										s = strings.TrimSpace(s)
										if s == "" {
											return "", nil
										}
										if len(s) > 180 {
											return s[:177] + "...", nil
										}
										return s, nil
									}),
									getters.Static("No description"),
								),
								Classes: "text-sm text-base-content/70",
							},
						},
					},
				},
			},
		},
	}
}

func targetOfInterestRelatedReportsSection() components.PageInterface {
	return &components.ContainerColumn{
		Page: components.Page{Key: "lacerate.TargetOfInterestDetailRelatedReports"},
		Children: []components.PageInterface{
			&components.DataTable[ReportPageData]{
				Page:     components.Page{Key: "lacerate.TargetOfInterestDetailRelatedReportsTable"},
				UID:      "lacerate-target-of-interest-related-reports-table",
				Title:    "Related reports",
				Subtitle: "Nearest embedding matches for this target.",
				Classes:  "w-full",
				Data:     getters.Key[components.ObjectList[ReportPageData]](ctxKeyRelatedReports),
				RowAttr: getters.RowAttrNavigate(
					lago.RoutePath("lacerate.ReportDetailRoute", map[string]getters.Getter[any]{
						"id": getters.Any(getters.Key[uint]("$row.Report.ID")),
					}),
				),
				Columns: []components.TableColumn{
					{
						Label: "Name",
						Children: []components.PageInterface{
							&components.FieldText{
								Getter: getters.IfOrElse(
									getters.Map(getters.Key[string]("$row.Report.Name"), func(_ context.Context, s string) (string, error) {
										return strings.TrimSpace(s), nil
									}),
									getters.Format("#%d", getters.Any(getters.Key[uint]("$row.Report.ID"))),
								),
							},
						},
					},
					{
						Label: "Kind",
						Children: []components.PageInterface{
							&components.FieldText{Getter: getters.Map(getters.Key[string]("$row.Report.Kind"), func(_ context.Context, s string) (string, error) {
								return reportKindLabel(s), nil
							})},
						},
					},
					{
						Label: "Details",
						Children: []components.PageInterface{
							&components.GetterPage{Getter: reportListConfigPageGetter()},
						},
					},
				},
			},
		},
	}
}

func targetOfInterestRelatedIntelSection() components.PageInterface {
	return &components.ContainerColumn{
		Page: components.Page{Key: "lacerate.TargetOfInterestDetailRelatedIntel"},
		Children: []components.PageInterface{
			&components.DataTable[Intel]{
				Page:     components.Page{Key: "lacerate.TargetOfInterestDetailRelatedIntelTable"},
				UID:      "lacerate-target-of-interest-related-intel-table",
				Title:    "Related intel",
				Subtitle: "Nearest embedding matches for this target.",
				Classes:  "w-full",
				Data:     getters.Key[components.ObjectList[Intel]](ctxKeyRelatedIntel),
				RowAttr: getters.RowAttrNavigate(
					lago.RoutePath("lacerate.IntelDetailRoute", map[string]getters.Getter[any]{
						"id": getters.Any(getters.Key[uint]("$row.ID")),
					}),
				),
				Columns: []components.TableColumn{
					{
						Label: "Source",
						Children: []components.PageInterface{
							&components.FieldText{
								Getter: getters.IfOrElse(
									getters.Map(getters.Key[string]("$row.Source.Name"), func(_ context.Context, s string) (string, error) {
										return strings.TrimSpace(s), nil
									}),
									getters.Static("—"),
								),
							},
						},
					},
					{
						Label: "Kind",
						Children: []components.PageInterface{
							&components.FieldText{
								Getter: getters.IfOrElse(
									getters.Map(getters.Key[string]("$row.Source.Kind"), func(_ context.Context, s string) (string, error) {
										return strings.TrimSpace(s), nil
									}),
									getters.Static("—"),
								),
							},
						},
					},
					{
						Label: "Datetime",
						Children: []components.PageInterface{
							&components.FieldDatetime{Getter: getters.Key[time.Time]("$row.Datetime")},
						},
					},
					{
						Label: "Content",
						Children: []components.PageInterface{
							&components.FieldText{Getter: intelContentPreviewCell()},
						},
					},
				},
			},
		},
	}
}
