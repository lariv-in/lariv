package p_lacerate

import (
	"context"
	"strings"
	"time"

	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	"maragu.dev/gomponents"
	html "maragu.dev/gomponents/html"
)

const ctxKeyWebsearchRelatedIntel = "websearchRelatedIntel"

func registerWebsearchPages() {
	registerWebsearchMenuPatch()
	registerWebsearchMenus()
	registerWebsearchList()
	registerWebsearchForms()
	registerWebsearchDetail()
}

func registerWebsearchMenuPatch() {
	lago.RegistryPage.Patch("lacerate.LacerateMenu", func(page components.PageInterface) components.PageInterface {
		menu, ok := page.(*components.SidebarMenu)
		if !ok {
			return page
		}
		menu.Children = append(menu.Children, &components.SidebarMenuItem{
			Title: getters.Static("Websearch"),
			Url:   lago.RoutePath("lacerate.WebsearchListRoute", nil),
		})
		return menu
	})
}

func registerWebsearchMenus() {
	lago.RegistryPage.Register("lacerate.WebsearchDetailMenu", &components.SidebarMenu{
		Title: getters.Format("Websearch #%d", getters.Any(getters.Key[uint]("websearch.ID"))),
		Back: &components.SidebarMenuItem{
			Title: getters.Static("Back to websearch"),
			Url:   lago.RoutePath("lacerate.WebsearchListRoute", nil),
		},
		Children: []components.PageInterface{
			&components.SidebarMenuItem{
				Title: getters.Static("Detail"),
				Url: lago.RoutePath("lacerate.WebsearchDetailRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("websearch.ID")),
				}),
			},
			&components.SidebarMenuItem{
				Title: getters.Static("Edit"),
				Url: lago.RoutePath("lacerate.WebsearchUpdateRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("websearch.ID")),
				}),
			},
		},
	})
}

func websearchQueryInput() components.PageInterface {
	return &components.ContainerColumn{
		Children: []components.PageInterface{
			&components.GetterPage{Getter: func(ctx context.Context) (components.PageInterface, error) {
				errs, ok := ctx.Value(getters.ContextKeyError).(map[string]error)
				if !ok || errs == nil || errs["_global"] == nil {
					return nil, nil
				}
				return &components.FieldText{
					Getter:  getters.Static(errs["_global"].Error()),
					Classes: "text-error",
				}, nil
			}},
			&components.ContainerError{
				Error: getters.Key[error]("$error.Query"),
				Children: []components.PageInterface{
					&components.InputText{
						Label:    "Search query",
						Name:     "Query",
						Required: true,
						Classes:  "w-full",
						Getter:   getters.Key[string]("$in.Query"),
					},
				},
			},
		},
	}
}

func websearchMainFormAttr() getters.Getter[gomponents.Node] {
	return func(ctx context.Context) (gomponents.Node, error) {
		base, err := getters.FormBubbling(getters.Static("lacerate.WebsearchMainForm"))(ctx)
		if err != nil {
			return nil, err
		}
		actionURL, err := lago.RoutePath("lacerate.WebsearchCreateRoute", nil)(ctx)
		if err != nil {
			return nil, err
		}
		return gomponents.Group{
			base,
			html.Action(actionURL),
		}, nil
	}
}

func registerWebsearchList() {
	lago.RegistryPage.Register("lacerate.WebsearchTable", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "lacerate.LacerateMenu"},
		},
		Children: []components.PageInterface{
			&components.FormListenBoostedPost{
				Name:      getters.Static("lacerate.WebsearchMainForm"),
				ActionURL: lago.RoutePath("lacerate.WebsearchCreateRoute", nil),
				Children: []components.PageInterface{
					&components.FormComponent[Websearch]{
						Attr:     websearchMainFormAttr(),
						Title:    "Websearch",
						Subtitle: "Search query creates intel and links it to this query record.",
						Classes:  "@container",
						ChildrenInput: []components.PageInterface{
							websearchQueryInput(),
						},
						ChildrenAction: []components.PageInterface{
							&components.ButtonSubmit{Label: "Search"},
						},
					},
				},
			},
			&components.DataTable[Websearch]{
				Page:    components.Page{Key: "lacerate.WebsearchTableBody"},
				UID:     "lacerate-websearch-table",
				Classes: "w-full",
				Data:    getters.Key[components.ObjectList[Websearch]]("websearches"),
				Actions: []components.PageInterface{
					&components.TableButtonCreate{Link: lago.RoutePath("lacerate.WebsearchCreateRoute", nil)},
				},
				RowAttr: getters.RowAttrNavigate(
					lago.RoutePath("lacerate.WebsearchDetailRoute", map[string]getters.Getter[any]{
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
						Label: "Query",
						Children: []components.PageInterface{
							&components.FieldText{
								Getter: getters.Map(getters.Key[string]("$row.Query"), func(_ context.Context, s string) (string, error) {
									return strings.TrimSpace(s), nil
								}),
							},
						},
					},
					{
						Label: "Created",
						Children: []components.PageInterface{
							&components.FieldDatetime{Getter: getters.Key[time.Time]("$row.CreatedAt")},
						},
					},
				},
			},
		},
	})
}

func registerWebsearchForms() {
	lago.RegistryPage.Register("lacerate.WebsearchCreateForm", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "lacerate.LacerateMenu"},
		},
		Children: []components.PageInterface{
			&components.FormListenBoostedPost{
				Name:      getters.Static("lacerate.WebsearchCreateForm"),
				ActionURL: lago.RoutePath("lacerate.WebsearchCreateRoute", nil),
				Children: []components.PageInterface{
					&components.FormComponent[Websearch]{
						Attr:     getters.FormBubbling(getters.Static("lacerate.WebsearchCreateForm")),
						Title:    "New websearch query",
						Subtitle: "Run websearch now and store linked intel.",
						Classes:  "@container",
						ChildrenInput: []components.PageInterface{
							websearchQueryInput(),
						},
						ChildrenAction: []components.PageInterface{
							&components.ButtonSubmit{Label: "Create"},
						},
					},
				},
			},
		},
	})

	lago.RegistryPage.Register("lacerate.WebsearchUpdateForm", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "lacerate.WebsearchDetailMenu"},
		},
		Children: []components.PageInterface{
			&components.FormListenBoostedPost{
				Name: getters.Static("lacerate.WebsearchUpdateForm"),
				ActionURL: lago.RoutePath("lacerate.WebsearchUpdateRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("websearch.ID")),
				}),
				Children: []components.PageInterface{
					&components.FormComponent[Websearch]{
						Getter:  getters.Key[Websearch]("websearch"),
						Attr:    getters.FormBubbling(getters.Static("lacerate.WebsearchUpdateForm")),
						Title:   "Edit websearch query",
						Classes: "@container",
						ChildrenInput: []components.PageInterface{
							websearchQueryInput(),
						},
						ChildrenAction: []components.PageInterface{
							&components.ContainerRow{
								Classes: "flex flex-wrap justify-between gap-2 mt-2 items-center",
								Children: []components.PageInterface{
									&components.ContainerRow{
										Classes: "flex justify-end gap-2",
										Children: []components.PageInterface{
											&components.ButtonSubmit{Label: "Save and run"},
											&components.ButtonModalForm{
												Label: "Delete",
												Icon:  "trash",
												Name:  getters.Static("lacerate.WebsearchDeleteForm"),
												Url: lago.RoutePath("lacerate.WebsearchDeleteRoute", map[string]getters.Getter[any]{
													"id": getters.Any(getters.Key[uint]("websearch.ID")),
												}),
												FormPostURL: lago.RoutePath("lacerate.WebsearchDeleteRoute", map[string]getters.Getter[any]{
													"id": getters.Any(getters.Key[uint]("websearch.ID")),
												}),
												ModalUID: "lacerate-websearch-delete-modal",
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

func websearchRelatedIntelSection() components.PageInterface {
	return &components.ContainerColumn{
		Page: components.Page{Key: "lacerate.WebsearchDetailRelatedIntel"},
		Children: []components.PageInterface{
			&components.ContainerRow{
				Classes: "flex justify-end mb-2",
				Children: []components.PageInterface{
					&components.ButtonModalForm{
						Label: "Delete related intel",
						Icon:  "trash",
						Name:  getters.Static("lacerate.WebsearchDeleteIntelForm"),
						Url: lago.RoutePath("lacerate.WebsearchDeleteIntelRoute", map[string]getters.Getter[any]{
							"id": getters.Any(getters.Key[uint]("websearch.ID")),
						}),
						FormPostURL: lago.RoutePath("lacerate.WebsearchDeleteIntelRoute", map[string]getters.Getter[any]{
							"id": getters.Any(getters.Key[uint]("websearch.ID")),
						}),
						ModalUID: "lacerate-websearch-delete-intel-modal",
						Classes:  "btn-error btn-sm",
					},
				},
			},
			&components.DataTable[Intel]{
				Page:     components.Page{Key: "lacerate.WebsearchDetailRelatedIntelTable"},
				UID:      "lacerate-websearch-related-intel-table",
				Title:    "Related intel",
				Subtitle: "Intel created by this query.",
				Classes:  "w-full",
				Data:     getters.Key[components.ObjectList[Intel]](ctxKeyWebsearchRelatedIntel),
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

func websearchStatusSection() components.PageInterface {
	return &components.ContainerColumn{
		Page: components.Page{Key: "lacerate.WebsearchDetailStatus"},
		Children: []components.PageInterface{
			&components.LabelInline{
				Title: "Status",
				Children: []components.PageInterface{
					&components.FieldText{Getter: getters.Key[string]("$in.Status")},
				},
			},
			&components.GetterPage{Getter: func(ctx context.Context) (components.PageInterface, error) {
				status, err := getters.Key[string]("$in.Status")(ctx)
				if err != nil {
					return nil, err
				}
				if status != "failed" {
					return nil, nil
				}
				msg, err := getters.Key[string]("$in.LastRunError")(ctx)
				if err != nil {
					return nil, nil
				}
				msg = strings.TrimSpace(msg)
				if msg == "" {
					return nil, nil
				}
				return &components.FieldText{
					Getter:  getters.Static(msg),
					Classes: "text-error text-sm",
				}, nil
			}},
		},
	}
}

func websearchDetailBody() components.PageInterface {
	return &components.GetterPage{Getter: func(ctx context.Context) (components.PageInterface, error) {
		status, err := getters.Key[string]("websearch.Status")(ctx)
		if err != nil {
			return nil, err
		}
		content := &components.ContainerColumn{
			Page: components.Page{Key: "lacerate.WebsearchDetailBody"},
			Children: []components.PageInterface{
				websearchStatusSection(),
				websearchRelatedIntelSection(),
			},
		}
		if status == "queued" || status == "running" {
			return &components.HTMXPolling{
				Page: components.Page{Key: "lacerate.WebsearchDetailPolling"},
				URL: lago.RoutePath("lacerate.WebsearchDetailRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("websearch.ID")),
				}),
				Children: []components.PageInterface{
					content,
				},
			}, nil
		}
		return content, nil
	}}
}

func registerWebsearchDetail() {
	lago.RegistryPage.Register("lacerate.WebsearchDetail", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "lacerate.WebsearchDetailMenu"},
		},
		Children: []components.PageInterface{
			&components.Detail[Websearch]{
				Getter: getters.Key[Websearch]("websearch"),
				Children: []components.PageInterface{
					&components.ContainerColumn{
						Page: components.Page{Key: "lacerate.WebsearchDetailContent"},
						Children: []components.PageInterface{
							&components.FieldTitle{Getter: getters.Format("Websearch #%d", getters.Any(getters.Key[uint]("$in.ID")))},
							&components.LabelInline{
								Title: "Query",
								Children: []components.PageInterface{
									&components.FieldText{Getter: getters.Key[string]("$in.Query")},
								},
							},
							websearchDetailBody(),
						},
					},
				},
			},
		},
	})

	lago.RegistryPage.Register("lacerate.WebsearchDeleteForm", &components.Modal{
		UID: "lacerate-websearch-delete-modal",
		Children: []components.PageInterface{
			&components.DeleteConfirmation{
				Title:   "Delete websearch query",
				Message: "Delete this websearch query and all links to its related intel?",
				Attr:    getters.FormBubbling(getters.Key[string]("$get.name")),
			},
		},
	})

	lago.RegistryPage.Register("lacerate.WebsearchDeleteIntelForm", &components.Modal{
		UID: "lacerate-websearch-delete-intel-modal",
		Children: []components.PageInterface{
			&components.DeleteConfirmation{
				Title:   "Delete related intel",
				Message: "Delete all intel linked to this websearch query? This cannot be undone.",
				Attr:    getters.FormBubbling(getters.Key[string]("$get.name")),
			},
		},
	})
}
