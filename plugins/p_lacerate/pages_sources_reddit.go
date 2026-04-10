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

func registerRedditSourcePages() {
	registerRedditSourceMenus()
	registerRedditSourceTable()
	registerRedditSourceForms()
	registerRedditSourceDetail()
	registerRedditSourceDelete()
}

func registerRedditSourceMenus() {
	lago.RegistryPage.Register("lacerate.LacerateMenu", &components.SidebarMenu{
		Title: getters.Static("lacerate"),
		Back: &components.SidebarMenuItem{
			Title: getters.Static("Back to All Apps"),
			Url:   lago.RoutePath("dashboard.AppsPage", nil),
		},
		Children: []components.PageInterface{
			&components.SidebarMenuItem{
				Title: getters.Static("Reddit sources"),
				Url:   lago.RoutePath("lacerate.DefaultRoute", nil),
			},
		},
	})

	lago.RegistryPage.Register("lacerate.RedditSourceDetailMenu", &components.SidebarMenu{
		Title: getters.Format("Reddit source: %s", getters.Any(getters.Key[string]("redditSource.Source.Name"))),
		Back: &components.SidebarMenuItem{
			Title: getters.Static("Back to Reddit sources"),
			Url:   lago.RoutePath("lacerate.DefaultRoute", nil),
		},
		Children: []components.PageInterface{
			&components.SidebarMenuItem{
				Title: getters.Static("Detail"),
				Url: lago.RoutePath("lacerate.DetailRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("redditSource.ID")),
				}),
			},
			&components.SidebarMenuItem{
				Title: getters.Static("Edit"),
				Url: lago.RoutePath("lacerate.UpdateRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("redditSource.ID")),
				}),
			},
		},
	})
}

func registerRedditSourceTable() {
	lago.RegistryPage.Register("lacerate.RedditSourcesTable", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "lacerate.LacerateMenu"},
		},
		Children: []components.PageInterface{
			&components.DataTable[RedditSource]{
				Page:    components.Page{Key: "lacerate.RedditSourcesTableBody"},
				UID:     "lacerate-reddit-sources-table",
				Classes: "w-full",
				Data:    getters.Key[components.ObjectList[RedditSource]]("reddit_sources"),
				Actions: []components.PageInterface{
					&components.TableButtonCreate{Link: lago.RoutePath("lacerate.CreateRoute", nil)},
				},
				RowAttr: getters.RowAttrNavigate(
					lago.RoutePath("lacerate.DetailRoute", map[string]getters.Getter[any]{
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
						Label: "Subreddits",
						Children: []components.PageInterface{
							&components.FieldText{Getter: getters.Map(getters.Key[datatypes.JSON]("$row.Subreddits"), func(_ context.Context, raw datatypes.JSON) (string, error) {
								if len(raw) == 0 {
									return "", nil
								}
								var items []string
								if err := json.Unmarshal(raw, &items); err != nil {
									return "", err
								}
								if len(items) == 0 {
									return "", nil
								}
								s := strings.Join(items, ", ")
								if len(s) > 120 {
									return s[:117] + "...", nil
								}
								return s, nil
							})},
						},
					},
					{
						Label: "Search query",
						Children: []components.PageInterface{
							&components.FieldText{Getter: getters.IfOrElse(
								getters.Map(getters.Key[string]("$row.SearchQuery"), func(_ context.Context, s string) (string, error) {
									s = strings.TrimSpace(s)
									if s == "" {
										return "", nil
									}
									if len(s) > 48 {
										return s[:45] + "...", nil
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

func redditSourceFormFields() components.ContainerColumn {
	return components.ContainerColumn{
		Page: components.Page{Key: "lacerate.RedditSourceFormFields"},
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
				Error: getters.Key[error]("$error.Subreddits"),
				Children: []components.PageInterface{
					&components.InputStringList{
						Label:   "Subreddits",
						Name:    "Subreddits",
						Classes: "w-full",
						Getter: getters.Key[datatypes.JSON]("$in.Subreddits"),
					},
				},
			},
			&components.ContainerError{
				Error: getters.Key[error]("$error.SearchQuery"),
				Children: []components.PageInterface{
					&components.InputText{
						Label:   "Search query",
						Name:    "SearchQuery",
						Getter:  getters.Key[string]("$in.SearchQuery"),
						Classes: "w-full",
					},
				},
			},
		},
	}
}

func registerRedditSourceForms() {
	lago.RegistryPage.Register("lacerate.RedditSourceCreateForm", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "lacerate.LacerateMenu"},
		},
		Children: []components.PageInterface{
			&components.FormListenBoostedPost{
				Name:      getters.Static("lacerate.RedditSourceCreateForm"),
				ActionURL: lago.RoutePath("lacerate.CreateRoute", nil),
				Children: []components.PageInterface{
					&components.FormComponent[RedditSource]{
						Attr:     getters.FormBubbling(getters.Static("lacerate.RedditSourceCreateForm")),
						Title:    "New Reddit source",
						Subtitle: "Name, optional minimum interval between runs (e.g. 1h, 30m), subreddits without r/, optional search query per subreddit.",
						Classes:  "@container",
						ChildrenInput: []components.PageInterface{
							redditSourceFormFields(),
						},
						ChildrenAction: []components.PageInterface{
							&components.ButtonSubmit{Label: "Create source"},
						},
					},
				},
			},
		},
	})

	lago.RegistryPage.Register("lacerate.RedditSourceUpdateForm", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "lacerate.RedditSourceDetailMenu"},
		},
		Children: []components.PageInterface{
			&components.FormListenBoostedPost{
				Name: getters.Static("lacerate.RedditSourceUpdateForm"),
				ActionURL: lago.RoutePath("lacerate.UpdateRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("redditSource.ID")),
				}),
				Children: []components.PageInterface{
					&components.FormComponent[RedditSource]{
						Getter:   getters.Key[RedditSource]("redditSource"),
						Attr:     getters.FormBubbling(getters.Static("lacerate.RedditSourceUpdateForm")),
						Title:    "Edit Reddit source",
						Subtitle: "Update name, duration, subreddits, and search query.",
						Classes:  "@container",
						ChildrenInput: []components.PageInterface{
							redditSourceFormFields(),
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
												Name:  getters.Static("lacerate.RedditSourceDeleteForm"),
												Url: lago.RoutePath("lacerate.DeleteRoute", map[string]getters.Getter[any]{
													"id": getters.Any(getters.Key[uint]("redditSource.ID")),
												}),
												FormPostURL: lago.RoutePath("lacerate.DeleteRoute", map[string]getters.Getter[any]{
													"id": getters.Any(getters.Key[uint]("redditSource.ID")),
												}),
												ModalUID: "lacerate-reddit-source-delete-modal",
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

func registerRedditSourceDetail() {
	lago.RegistryPage.Register("lacerate.RedditSourceDetail", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "lacerate.RedditSourceDetailMenu"},
		},
		Children: []components.PageInterface{
			&components.Detail[RedditSource]{
				Getter: getters.Key[RedditSource]("redditSource"),
				Children: []components.PageInterface{
					components.ContainerColumn{
						Page: components.Page{Key: "lacerate.RedditSourceDetailContent"},
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
								Title: "Subreddits",
								Children: []components.PageInterface{
									&components.FieldTextArea{
										Getter: getters.IfOrElse(
											getters.Map(getters.Key[datatypes.JSON]("$in.Subreddits"), func(_ context.Context, raw datatypes.JSON) (string, error) {
												if len(raw) == 0 {
													return "", nil
												}
												var items []string
												if err := json.Unmarshal(raw, &items); err != nil {
													return "", err
												}
												if len(items) == 0 {
													return "", nil
												}
												return strings.Join(items, "\n"), nil
											}),
											getters.Static("(none)"),
										),
										Classes: "text-sm",
									},
								},
							},
							&components.LabelInline{
								Title: "Search query",
								Children: []components.PageInterface{
									&components.FieldText{Getter: getters.IfOrElse(
										getters.Map(getters.Key[string]("$in.SearchQuery"), func(_ context.Context, s string) (string, error) {
											s = strings.TrimSpace(s)
											if s == "" {
												return "", nil
											}
											return s, nil
										}),
										getters.Static("(none — uses default subreddit listing)"),
									)},
								},
							},
						},
					},
				},
			},
		},
	})
}

func registerRedditSourceDelete() {
	lago.RegistryPage.Register("lacerate.RedditSourceDeleteForm", &components.Modal{
		UID: "lacerate-reddit-source-delete-modal",
		Children: []components.PageInterface{
			&components.DeleteConfirmation{
				Title:   "Delete Reddit source",
				Message: "Delete this Reddit source and its configuration? This cannot be undone.",
				Attr:    getters.FormBubbling(getters.Key[string]("$get.name")),
			},
		},
	})
}
