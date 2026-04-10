package p_lacerate

import (
	"context"
	"encoding/json"
	"log/slog"
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
		Title: getters.Format("Reddit source: %s", getters.Any(redditSourceDisplayNameGetter())),
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

func redditSourceDisplayNameGetter() getters.Getter[string] {
	return func(ctx context.Context) (string, error) {
		m, ok := ctx.Value(getters.ContextKeyIn).(map[string]any)
		if ok {
			if v, ok := m["Source.Name"]; ok {
				if s, ok := v.(string); ok && s != "" {
					return s, nil
				}
			}
		}
		return getters.Key[string]("redditSource.Source.Name")(ctx)
	}
}

func redditSourceNameInputGetter() getters.Getter[string] {
	return func(ctx context.Context) (string, error) {
		m, ok := ctx.Value(getters.ContextKeyIn).(map[string]any)
		if !ok {
			return "", nil
		}
		if v, ok := m["Source.Name"]; ok {
			if s, ok := v.(string); ok {
				return s, nil
			}
		}
		return "", nil
	}
}

func redditSourceDurationPtrGetter() getters.Getter[*time.Duration] {
	return func(ctx context.Context) (*time.Duration, error) {
		m, ok := ctx.Value(getters.ContextKeyIn).(map[string]any)
		if !ok {
			return nil, nil
		}
		if v, ok := m["Source.Duration"]; ok {
			switch d := v.(type) {
			case time.Duration:
				if d == 0 {
					return nil, nil
				}
				dup := d
				return &dup, nil
			case int64:
				if d == 0 {
					return nil, nil
				}
				dup := time.Duration(d)
				return &dup, nil
			case float64:
				if d == 0 {
					return nil, nil
				}
				dup := time.Duration(int64(d))
				return &dup, nil
			}
		}
		return nil, nil
	}
}

func redditSourceDurationCellGetter() getters.Getter[string] {
	return func(ctx context.Context) (string, error) {
		d, err := getters.Key[time.Duration]("$row.Source.Duration")(ctx)
		if err != nil {
			slog.Error("lacerate: reddit source duration cell getter", "error", err)
			return "", err
		}
		if d == 0 {
			return "—", nil
		}
		return d.String(), nil
	}
}

func redditSourceDurationDetailGetter() getters.Getter[string] {
	return func(ctx context.Context) (string, error) {
		d, err := getters.Key[time.Duration]("$in.Source.Duration")(ctx)
		if err != nil {
			slog.Error("lacerate: reddit source duration detail getter", "error", err)
			return "", err
		}
		if d == 0 {
			return "(not set)", nil
		}
		return d.String(), nil
	}
}

func redditSearchQueryInputGetter() getters.Getter[string] {
	return func(ctx context.Context) (string, error) {
		m, ok := ctx.Value(getters.ContextKeyIn).(map[string]any)
		if !ok {
			return "", nil
		}
		if v, ok := m["SearchQuery"]; ok {
			if s, ok := v.(string); ok {
				return s, nil
			}
		}
		return "", nil
	}
}

func redditSearchQueryCellGetter() getters.Getter[string] {
	return func(ctx context.Context) (string, error) {
		s, err := getters.Key[string]("$row.SearchQuery")(ctx)
		if err != nil {
			slog.Error("lacerate: reddit search query cell getter", "error", err)
			return "", err
		}
		s = strings.TrimSpace(s)
		if s == "" {
			return "—", nil
		}
		if len(s) > 48 {
			return s[:45] + "...", nil
		}
		return s, nil
	}
}

func redditSearchQueryDetailGetter() getters.Getter[string] {
	return func(ctx context.Context) (string, error) {
		s, err := getters.Key[string]("$in.SearchQuery")(ctx)
		if err != nil {
			slog.Error("lacerate: reddit search query detail getter", "error", err)
			return "", err
		}
		s = strings.TrimSpace(s)
		if s == "" {
			return "(none — uses default subreddit listing)", nil
		}
		return s, nil
	}
}

func redditSubredditsSummaryFromRowGetter() getters.Getter[string] {
	return func(ctx context.Context) (string, error) {
		raw, err := getters.Key[datatypes.JSON]("$row.Subreddits")(ctx)
		if err != nil || len(raw) == 0 {
			return "", nil
		}
		var items []string
		if err := json.Unmarshal(raw, &items); err != nil {
			slog.Error("lacerate: reddit subreddits summary unmarshal", "error", err)
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
	}
}

func redditSubredditsDetailGetter() getters.Getter[string] {
	return func(ctx context.Context) (string, error) {
		raw, err := getters.Key[datatypes.JSON]("$in.Subreddits")(ctx)
		if err != nil || len(raw) == 0 {
			return "(none)",
				nil
		}
		var items []string
		if err := json.Unmarshal(raw, &items); err != nil {
			slog.Error("lacerate: reddit subreddits detail unmarshal", "error", err)
			return "", err
		}
		if len(items) == 0 {
			return "(none)", nil
		}
		return strings.Join(items, "\n"), nil
	}
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
							&components.FieldText{Getter: redditSourceDurationCellGetter()},
						},
					},
					{
						Label: "Subreddits",
						Children: []components.PageInterface{
							&components.FieldText{Getter: redditSubredditsSummaryFromRowGetter()},
						},
					},
					{
						Label: "Search query",
						Children: []components.PageInterface{
							&components.FieldText{Getter: redditSearchQueryCellGetter()},
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
						Getter:   redditSourceNameInputGetter(),
					},
				},
			},
			&components.ContainerError{
				Error: getters.Key[error]("$error.Duration"),
				Children: []components.PageInterface{
					&components.InputDuration{
						Label:   "Duration",
						Name:    "Duration",
						Getter:  redditSourceDurationPtrGetter(),
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
						Getter: func(ctx context.Context) (datatypes.JSON, error) {
							m, ok := ctx.Value(getters.ContextKeyIn).(map[string]any)
							if !ok {
								return nil, nil
							}
							if v, ok := m["Subreddits"]; ok {
								if j, ok := v.(datatypes.JSON); ok {
									return j, nil
								}
							}
							return nil, nil
						},
					},
				},
			},
			&components.ContainerError{
				Error: getters.Key[error]("$error.SearchQuery"),
				Children: []components.PageInterface{
					&components.InputText{
						Label:   "Search query",
						Name:    "SearchQuery",
						Getter:  redditSearchQueryInputGetter(),
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
									&components.FieldText{Getter: redditSourceDurationDetailGetter()},
								},
							},
							&components.LabelInline{
								Title: "Subreddits",
								Children: []components.PageInterface{
									&components.FieldTextArea{
										Getter:  redditSubredditsDetailGetter(),
										Classes: "text-sm",
									},
								},
							},
							&components.LabelInline{
								Title: "Search query",
								Children: []components.PageInterface{
									&components.FieldText{Getter: redditSearchQueryDetailGetter()},
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
