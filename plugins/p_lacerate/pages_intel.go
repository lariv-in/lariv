package p_lacerate

import (
	"context"
	"log/slog"
	"strings"
	"time"

	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_filesystem"
)

func registerIntelPages() {
	registerIntellacerateMenuPatch()
	registerIntelMenus()
	registerIntelSourceSelectionPages()
	registerIntelTable()
	registerIntelForms()
	registerIntelDetail()
}

func registerIntellacerateMenuPatch() {
	lago.RegistryPage.Patch("lacerate.LacerateMenu", func(page components.PageInterface) components.PageInterface {
		menu, ok := page.(*components.SidebarMenu)
		if !ok {
			return page
		}
		menu.Children = append(menu.Children, &components.SidebarMenuItem{
			Title: getters.Static("Intel"),
			Url:   lago.RoutePath("lacerate.IntelListRoute", nil),
		})
		return menu
	})
}

func registerIntelMenus() {
	lago.RegistryPage.Register("lacerate.IntelDetailMenu", &components.SidebarMenu{
		Title: getters.Format("Intel #%d", getters.Any(getters.Key[uint]("intel.ID"))),
		Back: &components.SidebarMenuItem{
			Title: getters.Static("Back to Intel"),
			Url:   lago.RoutePath("lacerate.IntelListRoute", nil),
		},
		Children: []components.PageInterface{
			&components.SidebarMenuItem{
				Title: getters.Static("Detail"),
				Url: lago.RoutePath("lacerate.IntelDetailRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("intel.ID")),
				}),
			},
			&components.SidebarMenuItem{
				Title: getters.Static("Edit"),
				Url: lago.RoutePath("lacerate.IntelUpdateRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("intel.ID")),
				}),
			},
		},
	})
}

func intelContentPreviewCell() getters.Getter[string] {
	return func(ctx context.Context) (string, error) {
		s, err := getters.Key[string]("$row.Content")(ctx)
		if err != nil {
			slog.Error("lacerate: intel content preview cell", "error", err)
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

func intelFormFields() components.PageInterface {
	return &components.ContainerColumn{
		Page: components.Page{Key: "lacerate.IntelFormFields"},
		Children: []components.PageInterface{
			&components.ContainerError{
				Error: getters.Key[error]("$error.SourceID"),
				Children: []components.PageInterface{
					&components.InputForeignKey[Source]{
						Label:       "Source",
						Name:        "SourceID",
						Required:    true,
						Url:         lago.RoutePath("lacerate.SourceSelectRoute", nil),
						Display:     getters.Key[string]("$in.Source.Name"),
						Placeholder: "Select a source…",
						Getter:      getters.Association[Source](getters.Deref(getters.Key[*uint]("$in.SourceID"))),
					},
				},
			},
			&components.ContainerError{
				Error: getters.Key[error]("$error.Datetime"),
				Children: []components.PageInterface{
					&components.InputDatetime{
						Label:    "Datetime",
						Name:     "Datetime",
						Required: true,
						Getter:   getters.Key[time.Time]("$in.Datetime"),
					},
				},
			},
			&components.ContainerError{
				Error: getters.Key[error]("$error.Content"),
				Children: []components.PageInterface{
					&components.InputTextarea{
						Label:   "Content (markdown)",
						Name:    "Content",
						Rows:    14,
						Classes: "w-full font-mono text-sm",
						Getter:  getters.Key[string]("$in.Content"),
					},
				},
			},
			&components.ContainerError{
				Error: getters.Key[error]("$error.PreviewImageID"),
				Children: []components.PageInterface{
					&p_filesystem.InputVNode{
						Label:   "Preview image",
						Name:    "PreviewImageID",
						Classes: "w-full",
						VNode: func(ctx context.Context) (p_filesystem.VNode, error) {
							var zero p_filesystem.VNode
							if id, err := getters.Deref(getters.Key[*uint]("$in.PreviewImageID"))(ctx); err == nil && id != 0 {
								return getters.Association[p_filesystem.VNode](getters.Static(id))(ctx)
							}
							return zero, nil
						},
						AllowedFiletypes: []string{".jpg", ".jpeg", ".png", ".webp", ".gif"},
						Path: getters.IfOrElse(
							getters.Map(getters.Key[uint]("$in.ID"), func(ctx context.Context, id uint) (string, error) {
								if id == 0 {
									return "", nil
								}
								return getters.Format("lacerate/intel/%d", getters.Any(getters.Static(id)))(ctx)
							}),
							getters.Static("lacerate/intel/new"),
						),
					},
				},
			},
		},
	}
}

func registerIntelTable() {
	lago.RegistryPage.Register("lacerate.IntelTable", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "lacerate.LacerateMenu"},
		},
		Children: []components.PageInterface{
			&components.DataTable[Intel]{
				Page:    components.Page{Key: "lacerate.IntelTableBody"},
				UID:     "lacerate-intel-table",
				Classes: "w-full",
				Data:    getters.Key[components.ObjectList[Intel]]("intels"),
				Actions: []components.PageInterface{
					&components.TableButtonCreate{Link: lago.RoutePath("lacerate.IntelCreateRoute", nil)},
				},
				RowAttr: getters.RowAttrNavigate(
					lago.RoutePath("lacerate.IntelDetailRoute", map[string]getters.Getter[any]{
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
						Label: "Source",
						Children: []components.PageInterface{
							&components.FieldText{
								Getter: getters.Key[string]("$row.Source.Name"),
							},
						},
					},
					{
						Label: "Kind",
						Children: []components.PageInterface{
							&components.FieldText{Getter: getters.Key[string]("$row.Source.Kind")},
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
	})
}

func registerIntelForms() {
	createName := getters.Static("lacerate.IntelCreateForm")
	updateName := getters.Static("lacerate.IntelUpdateForm")
	deleteName := getters.Static("lacerate.IntelDeleteForm")

	lago.RegistryPage.Register("lacerate.IntelCreateForm", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "lacerate.LacerateMenu"},
		},
		Children: []components.PageInterface{
			&components.FormListenBoostedPost{
				Name:      createName,
				ActionURL: lago.RoutePath("lacerate.IntelCreateRoute", nil),
				Children: []components.PageInterface{
					&components.FormComponent[Intel]{
						Attr:     getters.FormBubbling(createName),
						Title:    "New Intel",
						Subtitle: "Link a source, set event datetime, add markdown content, and optionally attach a preview image.",
						Classes:  "@container",
						ChildrenInput: []components.PageInterface{
							intelFormFields(),
						},
						ChildrenAction: []components.PageInterface{
							&components.ButtonSubmit{Label: "Create"},
						},
					},
				},
			},
		},
	})

	lago.RegistryPage.Register("lacerate.IntelUpdateForm", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "lacerate.IntelDetailMenu"},
		},
		Children: []components.PageInterface{
			&components.FormListenBoostedPost{
				Name: updateName,
				ActionURL: lago.RoutePath("lacerate.IntelUpdateRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("intel.ID")),
				}),
				Children: []components.PageInterface{
					&components.FormComponent[Intel]{
						Getter:   getters.Key[Intel]("intel"),
						Attr:     getters.FormBubbling(updateName),
						Title:    "Edit Intel",
						Subtitle: "Update datetime, content, source, or preview image.",
						Classes:  "@container",
						ChildrenInput: []components.PageInterface{
							intelFormFields(),
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
												Url: lago.RoutePath("lacerate.IntelDeleteRoute", map[string]getters.Getter[any]{
													"id": getters.Any(getters.Key[uint]("intel.ID")),
												}),
												FormPostURL: lago.RoutePath("lacerate.IntelDeleteRoute", map[string]getters.Getter[any]{
													"id": getters.Any(getters.Key[uint]("intel.ID")),
												}),
												ModalUID: "lacerate-intel-delete-modal",
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

func registerIntelDetail() {
	lago.RegistryPage.Register("lacerate.IntelDetail", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "lacerate.IntelDetailMenu"},
		},
		Children: []components.PageInterface{
			&components.Detail[Intel]{
				Getter: getters.Key[Intel]("intel"),
				Children: []components.PageInterface{
					components.ContainerColumn{
						Page: components.Page{Key: "lacerate.IntelDetailContent"},
						Children: []components.PageInterface{
							&components.FieldTitle{
								Getter: getters.Format(
									"Intel — %s",
									getters.Any(getters.Key[string]("$in.Source.Name")),
								),
							},
							&components.FieldSubtitle{
								Getter: getters.Format(
									"%s · #%d",
									getters.Any(getters.Key[string]("$in.Source.Kind")),
									getters.Any(getters.Key[uint]("$in.ID")),
								),
							},
							&components.LabelInline{
								Title: "Datetime",
								Children: []components.PageInterface{
									&components.FieldDatetime{Getter: getters.Key[time.Time]("$in.Datetime")},
								},
							},
							&components.LabelInline{
								Title: "Preview",
								Children: []components.PageInterface{
									&p_filesystem.FieldPhoto{
										VNode:   getters.Association[p_filesystem.VNode](getters.Deref(getters.Key[*uint]("$in.PreviewImageID"))),
										Alt:     "Preview",
										Classes: "max-h-64 rounded border border-base-300",
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
							&components.FieldTitle{
								Getter:  getters.Static("Events"),
								Classes: "mt-8",
							},
							&components.DataTable[Event]{
								Page:     components.Page{Key: "lacerate.IntelDetailEventsTable"},
								UID:      "lacerate-intel-events-table",
								Subtitle: "Geocoded addresses for this intel; coordinates are stored for the map only.",
								Classes:  "w-full",
								Data:     getters.Key[components.ObjectList[Event]](ctxKeyIntelEvents),
								Columns: []components.TableColumn{
									{
										Label: "Datetime",
										Children: []components.PageInterface{
											&components.FieldDatetime{Getter: getters.Key[time.Time]("$row.Datetime")},
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
								Page:     components.Page{Key: "lacerate.IntelDetailRelatedTabs"},
								StateKey: "related_tab",
								Default:  getters.Static("Intel"),
								Tabs: map[string]getters.Getter[components.PageInterface]{
									"Targets": getters.Static[components.PageInterface](targetOfInterestRelatedSection()),
									"Reports": getters.Static[components.PageInterface](targetOfInterestRelatedReportsSection()),
									"Intel":   getters.Static[components.PageInterface](targetOfInterestRelatedIntelSection()),
								},
							},
						},
					},
				},
			},
		},
	})

	lago.RegistryPage.Register("lacerate.IntelDeleteForm", &components.Modal{
		UID: "lacerate-intel-delete-modal",
		Children: []components.PageInterface{
			&components.DeleteConfirmation{
				Title:   "Delete Intel",
				Message: "Delete this Intel row? This cannot be undone.",
				Attr:    getters.FormBubbling(getters.Key[string]("$get.name")),
			},
		},
	})
}

func registerIntelSourceSelectionPages() {
	lago.RegistryPage.Register("lacerate.SourceSelectionFilter", &components.FormComponent[Source]{
		Attr: getters.FormBoostedGet(lago.RoutePath("lacerate.SourceSelectRoute", nil)),
		ChildrenInput: []components.PageInterface{
			&components.InputText{
				Label:  "Name",
				Name:   "Name",
				Getter: getters.Key[string]("$get.Name"),
			},
			&components.InputText{
				Label:  "Kind",
				Name:   "Kind",
				Getter: getters.Key[string]("$get.Kind"),
			},
		},
		ChildrenAction: []components.PageInterface{
			components.ContainerRow{
				Classes: "flex gap-2",
				Children: []components.PageInterface{
					&components.ButtonSubmit{Label: "Apply"},
					&components.ButtonClear{Label: "Clear"},
				},
			},
		},
	})

	lago.RegistryPage.Register("lacerate.SourceSelectionTable", &components.Modal{
		UID: "lacerate-source-selection-modal",
		Children: []components.PageInterface{
			&components.DataTable[Source]{
				Page:  components.Page{Key: "lacerate.SourceSelectionTableBody"},
				UID:   "lacerate-source-selection-table",
				Title: "Select source",
				Data:  getters.Key[components.ObjectList[Source]]("sources"),
				RowAttr: getters.RowAttrSelect(
					"SourceID",
					getters.Key[uint]("$row.ID"),
					getters.Format(
						"%s (%s)",
						getters.Any(getters.Key[string]("$row.Name")),
						getters.Any(getters.Key[string]("$row.Kind")),
					),
				),
				Actions: []components.PageInterface{
					&components.TableButtonFilter{Child: lago.DynamicPage{Name: "lacerate.SourceSelectionFilter"}},
				},
				Columns: []components.TableColumn{
					{
						Label: "Name",
						Children: []components.PageInterface{
							&components.FieldText{Getter: getters.Key[string]("$row.Name")},
						},
					},
					{
						Label: "Kind",
						Children: []components.PageInterface{
							&components.FieldText{Getter: getters.Key[string]("$row.Kind")},
						},
					},
					{
						Label: "Duration",
						Children: []components.PageInterface{
							&components.FieldText{
								Getter: getters.IfOrElse(
									getters.Map(getters.Key[time.Duration]("$row.Duration"), func(ctx context.Context, d time.Duration) (string, error) {
										if d == 0 {
											return "", nil
										}
										return getters.Format("%v", getters.Any(getters.Static(d)))(ctx)
									}),
									getters.Static("—"),
								),
							},
						},
					},
				},
			},
		},
	})
}
