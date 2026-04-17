package p_lacerate

import (
	"context"
	"fmt"
	"time"

	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/registry"
	"maragu.dev/gomponents"
	html "maragu.dev/gomponents/html"
)

func registerSourcePages() {
	registerSourceMenus()
	registerSourceTable()
	registerSourceForms()
	registerSourceDetail()
	registerSourceDelete()
}

func registerSourceMenus() {
	lago.RegistryPage.Register("lacerate.LacerateMenu", &components.SidebarMenu{
		Title: getters.Static("lacerate"),
		Back: &components.SidebarMenuItem{
			Title: getters.Static("Back to All Apps"),
			Url:   lago.RoutePath("dashboard.AppsPage", nil),
		},
		Children: []components.PageInterface{
			&components.SidebarMenuItem{
				Title: getters.Static("Sources"),
				Url:   lago.RoutePath("lacerate.SourceListRoute", nil),
			},
		},
	})

	lago.RegistryPage.Register("lacerate.SourceDetailMenu", &components.SidebarMenu{
		Title: getters.Format("Source: %s", getters.Any(getters.Key[string]("sourcePageData.Source.Name"))),
		Back: &components.SidebarMenuItem{
			Title: getters.Static("Back to Sources"),
			Url:   lago.RoutePath("lacerate.SourceListRoute", nil),
		},
		Children: []components.PageInterface{
			&components.SidebarMenuItem{
				Title: getters.Static("Detail"),
				Url: lago.RoutePath("lacerate.SourceDetailRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("sourcePageData.Source.ID")),
				}),
			},
			&components.SidebarMenuItem{
				Title: getters.Static("Edit"),
				Url: lago.RoutePath("lacerate.SourceUpdateRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("sourcePageData.Source.ID")),
				}),
			},
		},
	})
}

var SourceListConfigPageMap = map[string]components.PageInterface{
	"reddit":               redditSourceListTableConfig(),
	"twitter":              twitterSourceListTableConfig(),
	"website":              websiteSourceListTableConfig(),
	sourceKindGoogleSearch: googleSearchSourceListTableConfig(),
	sourceKindWebsearch:    websearchSourceListTableConfig(),
	sourceKindDirectMedia:  directMediaSourceListTableConfig(),
}

func registerSourceTable() {
	lago.RegistryPage.Register("lacerate.SourcesTable", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "lacerate.LacerateMenu"},
		},
		Children: []components.PageInterface{
			&components.DataTable[SourcePageData]{
				Page:    components.Page{Key: "lacerate.SourcesTableBody"},
				UID:     "lacerate-sources-table",
				Classes: "w-full",
				Data:    getters.Key[components.ObjectList[SourcePageData]]("sources"),
				Actions: []components.PageInterface{
					&components.TableButtonCreate{Link: lago.RoutePath("lacerate.SourceCreateRoute", nil)},
				},
				RowAttr: getters.RowAttrNavigate(
					lago.RoutePath("lacerate.SourceDetailRoute", map[string]getters.Getter[any]{
						"id": getters.Any(getters.Key[uint]("$row.Source.ID")),
					}),
				),
				Columns: []components.TableColumn{
					{
						Label: "Name",
						Children: []components.PageInterface{
							&components.FieldText{Getter: getters.Key[string]("$row.Source.Name")},
						},
					},
					{
						Label: "Kind",
						Children: []components.PageInterface{
							&components.FieldText{Getter: getters.Map(getters.Key[string]("$row.Source.Kind"), func(_ context.Context, s string) (string, error) {
								if p, ok := registry.PairFromPairs(s, SourceKindChoices); ok {
									return p.Value, nil
								}
								return s, nil
							})},
						},
					},
					{
						Label: "Duration",
						Children: []components.PageInterface{
							&components.FieldText{Getter: getters.IfOrElse(
								getters.Map(getters.Key[time.Duration]("$row.Source.Duration"), func(_ context.Context, d time.Duration) (string, error) {
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
						Label: "Config",
						Children: []components.PageInterface{
							&components.GetterPage{Getter: func(ctx context.Context) (components.PageInterface, error) {
								current, err := getters.Key[string]("$row.Source.Kind")(ctx)
								if err != nil {
									return nil, err
								}
								page, ok := SourceListConfigPageMap[current]
								if !ok {
									return nil, fmt.Errorf("unknown source kind %q", current)
								}
								return page, nil
							}},
						},
					},
				},
			},
		},
	})
}

func sourceBaseFormFields() components.PageInterface {
	return &components.ContainerColumn{
		Page: components.Page{Key: "lacerate.SourceBaseFormFields"},
		Children: []components.PageInterface{
			&components.ContainerError{
				Error: getters.Key[error]("$error.Name"),
				Children: []components.PageInterface{
					&components.InputText{
						Label:    "Name",
						Name:     "Name",
						Required: true,
						Getter:   getters.Key[string]("$in.Source.Name"),
					},
				},
			},
			&components.ContainerError{
				Error: getters.Key[error]("$error.Duration"),
				Children: []components.PageInterface{
					&components.InputDuration{
						Label: "Duration",
						Name:  "Duration",
						Getter: getters.Map(getters.Key[time.Duration]("$in.Source.Duration"), func(_ context.Context, d time.Duration) (*time.Duration, error) {
							if d == 0 {
								return nil, nil
							}
							dup := d
							return &dup, nil
						}),
						Classes: "w-full",
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
						Choices:  getters.Static(SourceKindChoices),
						Getter: registry.PairGetter(
							getters.Key[string]("$in.Source.Kind"),
							getters.Static(registry.MapFromPairs(SourceKindChoices)),
						),
						Classes: "w-full",
						Attr:    getters.Static(gomponents.Attr("x-model", "kind")),
					},
				},
			},
		},
	}
}

func sourceKindFormMatch() components.PageInterface {
	redditFields := redditSourceFormFields()
	twitterFields := twitterSourceFormFields()
	websiteFields := websiteSourceFormFields()
	googleSearchFields := googleSearchSourceFormFields()
	websearchFields := websearchSourceFormFields()
	directMediaFields := directMediaSourceFormFields()
	return &components.ClientMatchIf{
		Key:   getters.Static("kind"),
		Match: getters.Static(map[string]components.PageInterface{"reddit": redditFields, "twitter": twitterFields, "website": websiteFields, sourceKindGoogleSearch: googleSearchFields, sourceKindWebsearch: websearchFields, sourceKindDirectMedia: directMediaFields}),
		Children: []components.PageInterface{
			redditFields,
			twitterFields,
			websiteFields,
			googleSearchFields,
			websearchFields,
			directMediaFields,
		},
	}
}

func sourceFormFields() components.PageInterface {
	// Single x-data scope: Kind select uses x-model="kind"; ClientMatchIf x-if reads same kind.
	return &components.ClientData{
		Page: components.Page{Key: "lacerate.SourceFormFields"},
		Data: `{ kind: '' }`,
		Init: `kind = (($el.querySelector('select[name=Kind]') || {}).value || '')`,
		Children: []components.PageInterface{
			&components.ContainerColumn{
				Children: []components.PageInterface{
					sourceBaseFormFields(),
					sourceKindFormMatch(),
				},
			},
		},
	}
}

func sourceFormActions(update bool) []components.PageInterface {
	if !update {
		return []components.PageInterface{
			&components.ButtonSubmit{Label: "Create source"},
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
							Name:  getters.Static("lacerate.SourceDeleteForm"),
							Url: lago.RoutePath("lacerate.SourceDeleteRoute", map[string]getters.Getter[any]{
								"id": getters.Any(getters.Key[uint]("sourcePageData.Source.ID")),
							}),
							FormPostURL: lago.RoutePath("lacerate.SourceDeleteRoute", map[string]getters.Getter[any]{
								"id": getters.Any(getters.Key[uint]("sourcePageData.Source.ID")),
							}),
							ModalUID: "lacerate-source-delete-modal",
							Classes:  "btn-error",
						},
					},
				},
			},
		},
	}
}

func registerSourceForms() {
	lago.RegistryPage.Register("lacerate.SourceCreateForm", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "lacerate.LacerateMenu"},
		},
		Children: []components.PageInterface{
			&components.FormListenBoostedPost{
				Name:      getters.Static("lacerate.SourceCreateForm"),
				ActionURL: lago.RoutePath("lacerate.SourceCreateRoute", nil),
				Children: []components.PageInterface{
					&components.FormComponent[SourcePageData]{
						Attr:     getters.FormBubbling(getters.Static("lacerate.SourceCreateForm")),
						Title:    "New source",
						Subtitle: "Choose source kind, name, optional poll interval, then fill in kind-specific fields.",
						Classes:  "@container",
						ChildrenInput: []components.PageInterface{
							sourceFormFields(),
						},
						ChildrenAction: sourceFormActions(false),
					},
				},
			},
		},
	})

	lago.RegistryPage.Register("lacerate.SourceUpdateForm", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "lacerate.SourceDetailMenu"},
		},
		Children: []components.PageInterface{
			&components.FormListenBoostedPost{
				Name: getters.Static("lacerate.SourceUpdateForm"),
				ActionURL: lago.RoutePath("lacerate.SourceUpdateRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("sourcePageData.Source.ID")),
				}),
				Children: []components.PageInterface{
					&components.FormComponent[SourcePageData]{
						Getter:   getters.Key[SourcePageData]("sourcePageData"),
						Attr:     getters.FormBubbling(getters.Static("lacerate.SourceUpdateForm")),
						Title:    "Edit source",
						Subtitle: "Update base source settings and kind-specific configuration.",
						Classes:  "@container",
						ChildrenInput: []components.PageInterface{
							sourceFormFields(),
						},
						ChildrenAction: sourceFormActions(true),
					},
				},
			},
		},
	})
}

func sourceDetailWorkerStatusGetter() getters.Getter[components.PageInterface] {
	return func(ctx context.Context) (components.PageInterface, error) {
		dur, err := getters.Key[time.Duration]("$in.Source.Duration")(ctx)
		if err != nil {
			return nil, err
		}
		id, err := getters.Key[uint]("$in.Source.ID")(ctx)
		if err != nil {
			return nil, err
		}
		active := SourceWorkerIsRunning(id)
		running, _ := SourceWorkerRunning(id)

		var text string
		if dur <= 0 {
			text = "Polling off (duration zero); Restart runs one fetch"
			if active && running {
				text = "Running (one fetch)"
			}
		} else {
			text = "Stopped (not polling)"
			if active {
				if running {
					text = "Running (fetching)"
				} else {
					text = "Waiting (between polls)"
				}
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

func sourceDetailWorkerActionsGetter() getters.Getter[components.PageInterface] {
	return func(ctx context.Context) (components.PageInterface, error) {
		id, err := getters.Key[uint]("$in.Source.ID")(ctx)
		if err != nil {
			return nil, err
		}
		active := SourceWorkerIsRunning(id)
		running, _ := SourceWorkerRunning(id)
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
			URL: lago.RoutePath("lacerate.SourceRestartWorkerRoute", map[string]getters.Getter[any]{
				"id": getters.Any(getters.Key[uint]("$in.Source.ID")),
			}),
			Icon:    "arrow-path",
			Classes: "btn-outline btn-primary btn-sm",
		}
		children := []components.PageInterface{restart}
		if active {
			children = append(children, &components.ButtonPost{
				Label: "Stop worker",
				URL: lago.RoutePath("lacerate.SourceStopWorkerRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("$in.Source.ID")),
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

func sourceDetailConfigPageGetter() getters.Getter[components.PageInterface] {
	return func(ctx context.Context) (components.PageInterface, error) {
		current, err := getters.Key[string]("$in.Source.Kind")(ctx)
		if err != nil {
			return nil, err
		}
		pageMap := map[string]components.PageInterface{
			"reddit":               redditSourceDetailFields(),
			"twitter":              twitterSourceDetailFields(),
			"website":              websiteSourceDetailFields(),
			sourceKindGoogleSearch: googleSearchSourceDetailFields(),
			sourceKindWebsearch:    websearchSourceDetailFields(),
			sourceKindDirectMedia:  directMediaSourceDetailFields(),
		}
		page, ok := pageMap[current]
		if !ok {
			return nil, fmt.Errorf("unknown source kind %q", current)
		}
		return page, nil
	}
}

func sourceDetailContent() components.PageInterface {
	return &components.ContainerColumn{
		Page: components.Page{Key: "lacerate.SourceDetailContent"},
		Children: []components.PageInterface{
			&components.FieldTitle{Getter: getters.Key[string]("$in.Source.Name")},
			&components.FieldSubtitle{Getter: getters.Map(getters.Key[string]("$in.Source.Kind"), func(_ context.Context, s string) (string, error) {
				if p, ok := registry.PairFromPairs(s, SourceKindChoices); ok {
					return p.Value, nil
				}
				return s, nil
			})},
			&components.GetterPage{Getter: sourceDetailWorkerStatusGetter()},
			&components.GetterPage{Getter: sourceDetailWorkerActionsGetter()},
			&components.LabelInline{
				Title: "Duration",
				Children: []components.PageInterface{
					&components.FieldText{Getter: getters.IfOrElse(
						getters.Map(getters.Key[time.Duration]("$in.Source.Duration"), func(_ context.Context, d time.Duration) (string, error) {
							if d == 0 {
								return "", nil
							}
							return d.String(), nil
						}),
						getters.Static("(not set)"),
					)},
				},
			},
			&components.GetterPage{Getter: sourceDetailConfigPageGetter()},
		},
	}
}

func sourceDetailBody() components.PageInterface {
	return &components.GetterPage{Getter: func(ctx context.Context) (components.PageInterface, error) {
		id, err := getters.Key[uint]("$in.Source.ID")(ctx)
		if err != nil {
			return nil, err
		}
		content := sourceDetailContent()
		if !SourceWorkerIsRunning(id) {
			return content, nil
		}
		url, err := lago.RoutePath("lacerate.SourceDetailRoute", map[string]getters.Getter[any]{
			"id": getters.Any(getters.Key[uint]("$in.Source.ID")),
		})(ctx)
		if err != nil {
			return nil, err
		}
		return &components.HTMXPolling{
			Page: components.Page{Key: "lacerate.SourceDetailPolling"},
			URL:  getters.Static(url),
			Children: []components.PageInterface{
				content,
			},
		}, nil
	}}
}

func registerSourceDetail() {
	lago.RegistryPage.Register("lacerate.SourceDetail", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "lacerate.SourceDetailMenu"},
		},
		Children: []components.PageInterface{
			&components.Detail[SourcePageData]{
				Getter: getters.Key[SourcePageData]("sourcePageData"),
				Children: []components.PageInterface{
					sourceDetailBody(),
				},
			},
		},
	})
}

func registerSourceDelete() {
	lago.RegistryPage.Register("lacerate.SourceDeleteForm", &components.Modal{
		UID: "lacerate-source-delete-modal",
		Children: []components.PageInterface{
			&components.DeleteConfirmation{
				Title:   "Delete source",
				Message: "Delete this source configuration? Existing intel will remain and lose its source link.",
				Attr:    getters.FormBubbling(getters.Key[string]("$get.name")),
			},
		},
	})
}
