package p_lacerate

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	"gorm.io/datatypes"
)

func registerTwitterSourcePages() {
	registerTwitterlacerateMenuPatch()
	registerTwitterSourceMenus()
	registerTwitterSourceTable()
	registerTwitterSourceForms()
	registerTwitterSourceDetail()
	registerTwitterSourceDelete()
}

func registerTwitterlacerateMenuPatch() {
	lago.RegistryPage.Patch("lacerate.LacerateMenu", func(page components.PageInterface) components.PageInterface {
		menu, ok := page.(*components.SidebarMenu)
		if !ok {
			return page
		}
		menu.Children = append(menu.Children, &components.SidebarMenuItem{
			Title: getters.Static("Twitter sources"),
			Url:   lago.RoutePath("lacerate.TwitterDefaultRoute", nil),
		})
		return menu
	})
}

func registerTwitterSourceMenus() {
	lago.RegistryPage.Register("lacerate.TwitterSourceDetailMenu", &components.SidebarMenu{
		Title: getters.Format("Twitter source: %s", getters.Any(getters.Key[string]("twitterSource.Source.Name"))),
		Back: &components.SidebarMenuItem{
			Title: getters.Static("Back to Twitter sources"),
			Url:   lago.RoutePath("lacerate.TwitterDefaultRoute", nil),
		},
		Children: []components.PageInterface{
			&components.SidebarMenuItem{
				Title: getters.Static("Detail"),
				Url: lago.RoutePath("lacerate.TwitterDetailRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("twitterSource.ID")),
				}),
			},
			&components.SidebarMenuItem{
				Title: getters.Static("Edit"),
				Url: lago.RoutePath("lacerate.TwitterUpdateRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("twitterSource.ID")),
				}),
			},
		},
	})
}

func registerTwitterSourceTable() {
	lago.RegistryPage.Register("lacerate.TwitterSourcesTable", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "lacerate.LacerateMenu"},
		},
		Children: []components.PageInterface{
			&components.DataTable[TwitterSource]{
				Page:    components.Page{Key: "lacerate.TwitterSourcesTableBody"},
				UID:     "lacerate-twitter-sources-table",
				Classes: "w-full",
				Data:    getters.Key[components.ObjectList[TwitterSource]]("twitter_sources"),
				Actions: []components.PageInterface{
					&components.TableButtonCreate{Link: lago.RoutePath("lacerate.TwitterCreateRoute", nil)},
				},
				RowAttr: getters.RowAttrNavigate(
					lago.RoutePath("lacerate.TwitterDetailRoute", map[string]getters.Getter[any]{
						"id": getters.Any(getters.Key[uint]("$row.ID")),
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
							&components.FieldText{Getter: getters.Key[string]("$row.Source.Kind")},
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
						Label: "Handles",
						Children: []components.PageInterface{
							&components.FieldText{Getter: getters.Map(getters.Key[datatypes.JSON]("$row.Handles"), func(_ context.Context, raw datatypes.JSON) (string, error) {
								if len(raw) == 0 {
									return "", nil
								}
								var items []string
								if err := json.Unmarshal(raw, &items); err != nil {
									return "", err
								}
								out := make([]string, 0, len(items))
								for _, s := range items {
									h := strings.TrimSpace(s)
									if h != "" {
										out = append(out, "@"+h)
									}
								}
								if len(out) == 0 {
									return "", nil
								}
								s := strings.Join(out, ", ")
								if len(s) > 120 {
									return s[:117] + "...", nil
								}
								return s, nil
							})},
						},
					},
				},
			},
		},
	})
}

func twitterSourceFormFields() components.ContainerColumn {
	return components.ContainerColumn{
		Page: components.Page{Key: "lacerate.TwitterSourceFormFields"},
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
						Label:   "Duration",
						Name:    "Duration",
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
				Error: getters.Key[error]("$error.Handles"),
				Children: []components.PageInterface{
					&InputTwitterHandleList{
						InputStringList: components.InputStringList{
							Label:   "Twitter handles",
							Name:    "Handles",
							Classes: "w-full",
							Getter: getters.Key[datatypes.JSON]("$in.Handles"),
						},
					},
				},
			},
		},
	}
}

func registerTwitterSourceForms() {
	lago.RegistryPage.Register("lacerate.TwitterSourceCreateForm", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "lacerate.LacerateMenu"},
		},
		Children: []components.PageInterface{
			&components.FormListenBoostedPost{
				Name:      getters.Static("lacerate.TwitterSourceCreateForm"),
				ActionURL: lago.RoutePath("lacerate.TwitterCreateRoute", nil),
				Children: []components.PageInterface{
					&components.FormComponent[TwitterSource]{
						Attr:     getters.FormBubbling(getters.Static("lacerate.TwitterSourceCreateForm")),
						Title:    "New Twitter source",
						Subtitle: "Name, optional poll interval, and handles (leading @ is stripped on save). Ingest mode is set globally in totschool.toml under [plugins.p_lacerate].",
						Classes:  "@container",
						ChildrenInput: []components.PageInterface{
							twitterSourceFormFields(),
						},
						ChildrenAction: []components.PageInterface{
							&components.ButtonSubmit{Label: "Create source"},
						},
					},
				},
			},
		},
	})

	lago.RegistryPage.Register("lacerate.TwitterSourceUpdateForm", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "lacerate.TwitterSourceDetailMenu"},
		},
		Children: []components.PageInterface{
			&components.FormListenBoostedPost{
				Name: getters.Static("lacerate.TwitterSourceUpdateForm"),
				ActionURL: lago.RoutePath("lacerate.TwitterUpdateRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("twitterSource.ID")),
				}),
				Children: []components.PageInterface{
					&components.FormComponent[TwitterSource]{
						Getter:  getters.Key[TwitterSource]("twitterSource"),
						Attr:    getters.FormBubbling(getters.Static("lacerate.TwitterSourceUpdateForm")),
						Title:   "Edit Twitter source",
						Classes: "@container",
						ChildrenInput: []components.PageInterface{
							twitterSourceFormFields(),
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
												Name:  getters.Static("lacerate.TwitterSourceDeleteForm"),
												Url: lago.RoutePath("lacerate.TwitterDeleteRoute", map[string]getters.Getter[any]{
													"id": getters.Any(getters.Key[uint]("twitterSource.ID")),
												}),
												FormPostURL: lago.RoutePath("lacerate.TwitterDeleteRoute", map[string]getters.Getter[any]{
													"id": getters.Any(getters.Key[uint]("twitterSource.ID")),
												}),
												ModalUID: "lacerate-twitter-source-delete-modal",
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

func registerTwitterSourceDetail() {
	lago.RegistryPage.Register("lacerate.TwitterSourceDetail", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "lacerate.TwitterSourceDetailMenu"},
		},
		Children: []components.PageInterface{
			&components.Detail[TwitterSource]{
				Getter: getters.Key[TwitterSource]("twitterSource"),
				Children: []components.PageInterface{
					components.ContainerColumn{
						Page: components.Page{Key: "lacerate.TwitterSourceDetailContent"},
						Children: []components.PageInterface{
							&components.FieldTitle{Getter: getters.Key[string]("$in.Source.Name")},
							&components.FieldSubtitle{Getter: getters.Key[string]("$in.Source.Kind")},
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
							&components.LabelInline{
								Title: "Handles",
								Children: []components.PageInterface{
									&components.FieldTextArea{
										Getter: getters.IfOrElse(
											getters.Map(getters.Key[datatypes.JSON]("$in.Handles"), func(_ context.Context, raw datatypes.JSON) (string, error) {
												if len(raw) == 0 {
													return "", nil
												}
												var items []string
												if err := json.Unmarshal(raw, &items); err != nil {
													return "", err
												}
												var lines []string
												for _, s := range items {
													h := strings.TrimSpace(s)
													if h != "" {
														lines = append(lines, "@"+h)
													}
												}
												if len(lines) == 0 {
													return "", nil
												}
												return strings.Join(lines, "\n"), nil
											}),
											getters.Static("(none)"),
										),
										Classes: "text-sm",
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

func registerTwitterSourceDelete() {
	lago.RegistryPage.Register("lacerate.TwitterSourceDeleteForm", &components.Modal{
		UID: "lacerate-twitter-source-delete-modal",
		Children: []components.PageInterface{
			&components.DeleteConfirmation{
				Title:   "Delete Twitter source",
				Message: "Delete this Twitter source and its configuration? This cannot be undone.",
				Attr:    getters.FormBubbling(getters.Key[string]("$get.name")),
			},
		},
	})
}
