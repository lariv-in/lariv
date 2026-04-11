package p_lacerate

import (
	"context"
	"log/slog"
	"strings"
	"time"

	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	"maragu.dev/gomponents"
	html "maragu.dev/gomponents/html"
)

func registerLookupPages() {
	registerLookuplacerateMenuPatch()
	registerLookupMenus()
	registerLookupTable()
	registerLookupForms()
	registerLookupDetail()
}

func registerLookuplacerateMenuPatch() {
	lago.RegistryPage.Patch("lacerate.LacerateMenu", func(page components.PageInterface) components.PageInterface {
		menu, ok := page.(*components.SidebarMenu)
		if !ok {
			return page
		}
		menu.Children = append(menu.Children, &components.SidebarMenuItem{
			Title: getters.Static("Lookups"),
			Url:   lago.RoutePath("lacerate.LookupListRoute", nil),
		})
		return menu
	})
}

func registerLookupMenus() {
	lago.RegistryPage.Register("lacerate.LookupDetailMenu", &components.SidebarMenu{
		Title: getters.Format("Lookup #%d", getters.Any(getters.Key[uint]("lookup.ID"))),
		Back: &components.SidebarMenuItem{
			Title: getters.Static("Back to lookups"),
			Url:   lago.RoutePath("lacerate.LookupListRoute", nil),
		},
		Children: []components.PageInterface{
			&components.SidebarMenuItem{
				Title: getters.Static("Detail"),
				Url: lago.RoutePath("lacerate.LookupDetailRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("lookup.ID")),
				}),
			},
			&components.SidebarMenuItem{
				Title: getters.Static("Edit"),
				Url: lago.RoutePath("lacerate.LookupUpdateRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("lookup.ID")),
				}),
			},
		},
	})
}

func lookupContentPreviewCell() getters.Getter[string] {
	return func(ctx context.Context) (string, error) {
		s, err := getters.Key[string]("$row.Content")(ctx)
		if err != nil {
			slog.Error("lacerate: lookup content preview cell", "error", err)
			return "", err
		}
		s = strings.TrimSpace(s)
		if s == "" {
			return "—", nil
		}
		if len(s) > 120 {
			return s[:117] + "...", nil
		}
		return s, nil
	}
}

func registerLookupTable() {
	lago.RegistryPage.Register("lacerate.LookupsTable", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "lacerate.LacerateMenu"},
		},
		Children: []components.PageInterface{
			&components.DataTable[Lookup]{
				Page:    components.Page{Key: "lacerate.LookupsTableBody"},
				UID:     "lacerate-lookups-table",
				Classes: "w-full",
				Data:    getters.Key[components.ObjectList[Lookup]]("lookups"),
				Actions: []components.PageInterface{
					&components.TableButtonCreate{Link: lago.RoutePath("lacerate.LookupCreateRoute", nil)},
				},
				RowAttr: getters.RowAttrNavigate(
					lago.RoutePath("lacerate.LookupDetailRoute", map[string]getters.Getter[any]{
						"id": getters.Any(getters.Key[uint]("$row.ID")),
					}),
				),
				Columns: []components.TableColumn{
					{
						Label: "ID",
						Children: []components.PageInterface{
							&components.FieldText{Getter: getters.Format("%d", getters.Any(getters.Key[uint]("$row.ID")))},
						},
					},
					{
						Label: "Update interval",
						Children: []components.PageInterface{
							&components.FieldText{Getter: getters.IfOrElse(
								getters.Map(getters.Deref(getters.Key[*time.Duration]("$row.UpdateInterval")), func(_ context.Context, d time.Duration) (string, error) {
									if d == 0 {
										return "", nil
									}
									return d.String(), nil
								}),
								getters.Static("—"),
							)},
						},
					},
					{
						Label: "Content",
						Children: []components.PageInterface{
							&components.FieldText{Getter: lookupContentPreviewCell()},
						},
					},
				},
			},
		},
	})
}

func lookupFormFields() components.PageInterface {
	return &components.ContainerColumn{
		Page: components.Page{Key: "lacerate.LookupFormFields"},
		Children: []components.PageInterface{
			&components.ContainerError{
				Error: getters.Key[error]("$error.UpdateInterval"),
				Children: []components.PageInterface{
					&components.InputDuration{
						Label:   "Update interval (optional)",
						Name:    "UpdateInterval",
						Getter:  getters.Key[*time.Duration]("$in.UpdateInterval"),
						Classes: "w-full",
					},
				},
			},
			&components.ContainerError{
				Error: getters.Key[error]("$error.Content"),
				Children: []components.PageInterface{
					&components.InputTextarea{
						Label:   "Content",
						Name:    "Content",
						Rows:    16,
						Classes: "w-full font-mono text-sm",
						Getter:  getters.Key[string]("$in.Content"),
					},
				},
			},
		},
	}
}

func registerLookupForms() {
	createName := getters.Static("lacerate.LookupCreateForm")
	updateName := getters.Static("lacerate.LookupUpdateForm")
	deleteName := getters.Static("lacerate.LookupDeleteForm")

	lago.RegistryPage.Register("lacerate.LookupCreateForm", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "lacerate.LacerateMenu"},
		},
		Children: []components.PageInterface{
			&components.FormListenBoostedPost{
				Name:      createName,
				ActionURL: lago.RoutePath("lacerate.LookupCreateRoute", nil),
				Children: []components.PageInterface{
					&components.FormComponent[Lookup]{
						Attr:     getters.FormBubbling(createName),
						Title:    "New lookup",
						Subtitle: "Arbitrary text content stored for reference.",
						Classes:  "@container",
						ChildrenInput: []components.PageInterface{
							lookupFormFields(),
						},
						ChildrenAction: []components.PageInterface{
							&components.ButtonSubmit{Label: "Create"},
						},
					},
				},
			},
		},
	})

	lago.RegistryPage.Register("lacerate.LookupUpdateForm", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "lacerate.LookupDetailMenu"},
		},
		Children: []components.PageInterface{
			&components.FormListenBoostedPost{
				Name: updateName,
				ActionURL: lago.RoutePath("lacerate.LookupUpdateRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("lookup.ID")),
				}),
				Children: []components.PageInterface{
					&components.FormComponent[Lookup]{
						Getter:  getters.Key[Lookup]("lookup"),
						Attr:    getters.FormBubbling(updateName),
						Title:   "Edit lookup",
						Classes: "@container",
						ChildrenInput: []components.PageInterface{
							lookupFormFields(),
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
												Url: lago.RoutePath("lacerate.LookupDeleteRoute", map[string]getters.Getter[any]{
													"id": getters.Any(getters.Key[uint]("lookup.ID")),
												}),
												FormPostURL: lago.RoutePath("lacerate.LookupDeleteRoute", map[string]getters.Getter[any]{
													"id": getters.Any(getters.Key[uint]("lookup.ID")),
												}),
												ModalUID: "lacerate-lookup-delete-modal",
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

func lookupDetailWorkerStatusGetter() getters.Getter[components.PageInterface] {
	return func(ctx context.Context) (components.PageInterface, error) {
		intervalPtr, err := getters.Key[*time.Duration]("$in.UpdateInterval")(ctx)
		if err != nil {
			return nil, err
		}
		if intervalPtr == nil || *intervalPtr <= 0 {
			return nil, nil
		}
		id, err := getters.Key[uint]("$in.ID")(ctx)
		if err != nil {
			return nil, err
		}
		active := LookupWorkerIsRunning(id)
		text := "Stopped (not polling)"
		if active {
			if running, _ := LookupWorkerRunning(id); running {
				text = "Running (lookup agent)"
			} else {
				text = "Waiting (between runs)"
			}
		}
		return &components.LabelInline{
			Title: "Worker",
			Children: []components.PageInterface{
				&components.FieldText{Getter: getters.Static(text)},
			},
		}, nil
	}
}

func lookupDetailWorkerActionsGetter() getters.Getter[components.PageInterface] {
	return func(ctx context.Context) (components.PageInterface, error) {
		id, err := getters.Key[uint]("$in.ID")(ctx)
		if err != nil {
			return nil, err
		}
		active := LookupWorkerIsRunning(id)
		running, _ := LookupWorkerRunning(id)
		label := "Restart worker"
		var attr getters.Getter[gomponents.Node]
		if running {
			attr = func(context.Context) (gomponents.Node, error) {
				return gomponents.Group{html.Disabled(), html.Class("btn-disabled")}, nil
			}
		}
		restart := &components.ButtonPost{
			Label: label,
			Attr:  attr,
			URL: lago.RoutePath("lacerate.LookupRestartWorkerRoute", map[string]getters.Getter[any]{
				"id": getters.Any(getters.Key[uint]("$in.ID")),
			}),
			Icon:    "arrow-path",
			Classes: "btn-outline btn-primary btn-sm",
		}
		children := []components.PageInterface{restart}
		if active {
			children = append(children, &components.ButtonPost{
				Label: "Stop worker",
				URL: lago.RoutePath("lacerate.LookupStopWorkerRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("$in.ID")),
				}),
				Icon:    "stop",
				Classes: "btn-outline btn-error btn-sm",
			})
		}
		return &components.ContainerRow{
			Classes:  "flex flex-wrap gap-2 items-center mb-2",
			Children: children,
		}, nil
	}
}

func registerLookupDetail() {
	lago.RegistryPage.Register("lacerate.LookupDetail", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "lacerate.LookupDetailMenu"},
		},
		Children: []components.PageInterface{
			&components.Detail[Lookup]{
				Getter: getters.Key[Lookup]("lookup"),
				Children: []components.PageInterface{
					components.ContainerColumn{
						Page: components.Page{Key: "lacerate.LookupDetailContent"},
						Children: []components.PageInterface{
							&components.FieldTitle{
								Getter: getters.Format("Lookup #%d", getters.Any(getters.Key[uint]("$in.ID"))),
							},
							&components.GetterPage{Getter: lookupDetailWorkerStatusGetter()},
							&components.GetterPage{Getter: lookupDetailWorkerActionsGetter()},
							&components.LabelInline{
								Title: "Update interval",
								Children: []components.PageInterface{
									&components.FieldText{Getter: getters.IfOrElse(
										getters.Map(getters.Deref(getters.Key[*time.Duration]("$in.UpdateInterval")), func(_ context.Context, d time.Duration) (string, error) {
											if d == 0 {
												return "", nil
											}
											return d.String(), nil
										}),
										getters.Static("(not set)"),
									)},
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
							lookupDetailTouchedTargetsOfInterestSection(),
							lookupDetailLogSection(),
						},
					},
				},
			},
		},
	})

	lago.RegistryPage.Register("lacerate.LookupDeleteForm", &components.Modal{
		UID: "lacerate-lookup-delete-modal",
		Children: []components.PageInterface{
			&components.DeleteConfirmation{
				Title:   "Delete lookup",
				Message: "Delete this lookup? This cannot be undone.",
				Attr:    getters.FormBubbling(getters.Key[string]("$get.name")),
			},
		},
	})
}
