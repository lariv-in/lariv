package p_lacerate

import (
	"context"
	"maps"
	"time"

	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/registry"
	"maragu.dev/gomponents"
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
		Title: getters.Format("Report — %s", getters.Any(getters.Key[string]("reportPageData.Report.Name"))),
		Back: &components.SidebarMenuItem{
			Title: getters.Static("Back to Reports"),
			Url:   lago.RoutePath("lacerate.ReportListRoute", nil),
		},
		Children: []components.PageInterface{
			&components.SidebarMenuItem{
				Title: getters.Static("Detail"),
				Url: lago.RoutePath("lacerate.ReportDetailRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("reportPageData.Report.ID")),
				}),
			},
			&components.SidebarMenuItem{
				Title: getters.Static("Edit"),
				Url: lago.RoutePath("lacerate.ReportUpdateRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("reportPageData.Report.ID")),
				}),
			},
		},
	})
}

func reportNameGetter() getters.Getter[string] {
	return func(ctx context.Context) (string, error) {
		if s, err := getters.Key[string]("$in.Name")(ctx); err == nil {
			return s, nil
		}
		if s, err := getters.Key[string]("$in.Report.Name")(ctx); err == nil {
			return s, nil
		}
		return "", nil
	}
}

func reportDescriptionGetter() getters.Getter[string] {
	return func(ctx context.Context) (string, error) {
		if s, err := getters.Key[string]("$in.Description")(ctx); err == nil {
			return s, nil
		}
		if s, err := getters.Key[string]("$in.Report.Description")(ctx); err == nil {
			return s, nil
		}
		return "", nil
	}
}

func reportKindGetter() getters.Getter[registry.Pair[string, string]] {
	return func(ctx context.Context) (registry.Pair[string, string], error) {
		s, err := getters.Key[string]("$in.Kind")(ctx)
		if err != nil || s == "" {
			s, _ = getters.Key[string]("$in.Report.Kind")(ctx)
		}
		if s == "" {
			return registry.Pair[string, string]{}, nil
		}
		if p, ok := registry.PairFromPairs(s, ReportKindChoices); ok {
			return p, nil
		}
		return registry.Pair[string, string]{Key: s, Value: s}, nil
	}
}

func reportFormContextGetter() getters.Getter[map[string]any] {
	return func(ctx context.Context) (map[string]any, error) {
		data, err := getters.Key[ReportPageData]("reportPageData")(ctx)
		if err != nil {
			return nil, err
		}
		tz, _ := ctx.Value("$tz").(*time.Location)
		out := map[string]any{
			"Name":        data.Report.Name,
			"Description": data.Report.Description,
			"Kind":        data.Report.Kind,
			"Report":      data.Report,
			"Briefing":    data.Briefing,
			"Timeline":    data.Timeline,
		}
		if data.Briefing != nil {
			out["BriefingContent"] = data.Briefing.Content
		}
		if data.Timeline != nil {
			out["TimelineEntriesJSON"] = timelineEntriesJSONValue(data.Timeline.Entries, tz)
		}
		if current, ok := ctx.Value(getters.ContextKeyIn).(map[string]any); ok {
			maps.Copy(out, current)
		}
		return out, nil
	}
}

func reportBaseFormFields() components.PageInterface {
	return &components.ContainerColumn{
		Page: components.Page{Key: "lacerate.ReportBaseFormFields"},
		Children: []components.PageInterface{
			&components.ContainerError{
				Error: getters.Key[error]("$error.Name"),
				Children: []components.PageInterface{
					&components.InputText{
						Label:    "Name",
						Name:     "Name",
						Required: true,
						Getter:   reportNameGetter(),
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
						Getter:  reportDescriptionGetter(),
					},
				},
			},
			&components.ContainerError{
				Error: getters.Key[error]("$error.Kind"),
				Children: []components.PageInterface{
					&components.InputSelect[string]{
						Label:    "Kind",
						Name:     "Kind",
						Required: true,
						Choices:  getters.Static(ReportKindChoices),
						Getter:   reportKindGetter(),
						Classes:  "w-full",
						Attr:     getters.Static(gomponents.Attr("x-model", "kind")),
					},
				},
			},
		},
	}
}

func reportKindFormMatch() components.PageInterface {
	briefingFields := briefingReportFormFields()
	timelineFields := timelineReportFormFields()
	return &components.ClientMatchIf{
		Key:   getters.Static("kind"),
		Match: getters.Static(map[string]components.PageInterface{"briefing": briefingFields, "timeline": timelineFields}),
		Children: []components.PageInterface{
			briefingFields,
			timelineFields,
		},
	}
}

func reportFormFields() components.PageInterface {
	return &components.ClientData{
		Page: components.Page{Key: "lacerate.ReportFormFields"},
		Data: `{ kind: '' }`,
		Init: `kind = (($el.querySelector('select[name=Kind]') || {}).value || '')`,
		Children: []components.PageInterface{
			&components.ContainerColumn{
				Children: []components.PageInterface{
					reportBaseFormFields(),
					reportKindFormMatch(),
				},
			},
		},
	}
}

var ReportListConfigPageMap = map[string]components.PageInterface{
	"briefing": briefingReportListTableConfig(),
	"timeline": timelineReportListTableConfig(),
}

func reportListConfigPageGetter() getters.Getter[components.PageInterface] {
	return func(ctx context.Context) (components.PageInterface, error) {
		current, err := getters.Key[string]("$row.Report.Kind")(ctx)
		if err != nil {
			return nil, err
		}
		page, ok := ReportListConfigPageMap[current]
		if !ok {
			return nil, nil
		}
		return page, nil
	}
}

func registerReportTable() {
	lago.RegistryPage.Register("lacerate.ReportsTable", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "lacerate.LacerateMenu"},
		},
		Children: []components.PageInterface{
			&components.DataTable[ReportPageData]{
				Page:    components.Page{Key: "lacerate.ReportsTableBody"},
				UID:     "lacerate-reports-table",
				Classes: "w-full",
				Data:    getters.Key[components.ObjectList[ReportPageData]]("reports"),
				Actions: []components.PageInterface{
					&components.TableButtonCreate{Link: lago.RoutePath("lacerate.ReportCreateRoute", nil)},
				},
				RowAttr: getters.RowAttrNavigate(
					lago.RoutePath("lacerate.ReportDetailRoute", map[string]getters.Getter[any]{
						"id": getters.Any(getters.Key[uint]("$row.Report.ID")),
					}),
				),
				Columns: []components.TableColumn{
					{
						Label: "Name",
						Children: []components.PageInterface{
							&components.FieldText{Getter: getters.Key[string]("$row.Report.Name")},
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
	})
}

func reportFormActions(update bool) []components.PageInterface {
	if !update {
		return []components.PageInterface{
			&components.ButtonSubmit{Label: "Create report"},
		}
	}
	return []components.PageInterface{
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
							Name:  getters.Static("lacerate.ReportDeleteForm"),
							Url: lago.RoutePath("lacerate.ReportDeleteRoute", map[string]getters.Getter[any]{
								"id": getters.Any(getters.Key[uint]("reportPageData.Report.ID")),
							}),
							FormPostURL: lago.RoutePath("lacerate.ReportDeleteRoute", map[string]getters.Getter[any]{
								"id": getters.Any(getters.Key[uint]("reportPageData.Report.ID")),
							}),
							ModalUID: "lacerate-report-delete-modal",
							Classes:  "btn-error",
						},
					},
				},
			},
		},
	}
}

func registerReportForms() {
	createName := getters.Static("lacerate.ReportCreateForm")
	updateName := getters.Static("lacerate.ReportUpdateForm")

	lago.RegistryPage.Register("lacerate.ReportCreateForm", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "lacerate.LacerateMenu"},
		},
		Children: []components.PageInterface{
			&components.FormListenBoostedPost{
				Name:      createName,
				ActionURL: lago.RoutePath("lacerate.ReportCreateRoute", nil),
				Children: []components.PageInterface{
					&components.FormComponent[ReportPageData]{
						Attr:     getters.FormBubbling(createName),
						Title:    "New report",
						Subtitle: "Choose report kind, add shared metadata, then fill kind-specific content.",
						Classes:  "@container",
						ChildrenInput: []components.PageInterface{
							reportFormFields(),
						},
						ChildrenAction: reportFormActions(false),
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
					"id": getters.Any(getters.Key[uint]("reportPageData.Report.ID")),
				}),
				Children: []components.PageInterface{
					&components.FormComponent[map[string]any]{
						Getter:   reportFormContextGetter(),
						Attr:     getters.FormBubbling(updateName),
						Title:    "Edit report",
						Subtitle: "Update shared metadata and kind-specific report content.",
						Classes:  "@container",
						ChildrenInput: []components.PageInterface{
							reportFormFields(),
						},
						ChildrenAction: reportFormActions(true),
					},
				},
			},
		},
	})
}

func reportDetailConfigPageGetter() getters.Getter[components.PageInterface] {
	return func(ctx context.Context) (components.PageInterface, error) {
		current, err := getters.Key[string]("$in.Report.Kind")(ctx)
		if err != nil {
			return nil, err
		}
		pageMap := map[string]components.PageInterface{
			"briefing": briefingReportDetailFields(),
			"timeline": timelineReportDetailFields(),
		}
		page, ok := pageMap[current]
		if !ok {
			return nil, nil
		}
		return page, nil
	}
}

func registerReportDetail() {
	lago.RegistryPage.Register("lacerate.ReportDetail", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "lacerate.ReportDetailMenu"},
		},
		Children: []components.PageInterface{
			&components.Detail[ReportPageData]{
				Getter: getters.Key[ReportPageData]("reportPageData"),
				Children: []components.PageInterface{
					components.ContainerColumn{
						Page: components.Page{Key: "lacerate.ReportDetailContent"},
						Children: []components.PageInterface{
							&components.FieldTitle{Getter: getters.Key[string]("$in.Report.Name")},
							&components.FieldSubtitle{Getter: getters.Map(getters.Key[string]("$in.Report.Kind"), func(_ context.Context, s string) (string, error) {
								return reportKindLabel(s), nil
							})},
							&components.LabelInline{
								Title: "Description",
								Children: []components.PageInterface{
									&components.FieldMarkdown{
										Getter:  getters.Key[string]("$in.Report.Description"),
										Classes: "prose prose-sm max-w-none",
									},
								},
							},
							&components.GetterPage{Getter: reportDetailConfigPageGetter()},
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
