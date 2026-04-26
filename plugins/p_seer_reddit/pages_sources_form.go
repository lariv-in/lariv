package p_seer_reddit

import (
	"context"

	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	"gorm.io/datatypes"
)

var redditSourceCreateFormDefaults = RedditSource{
	Subreddits:    datatypes.JSON([]byte("[]")),
	MaxFreshPosts: defaultMaxFreshPosts,
}

func redditSourceCreateFormFields() components.PageInterface {
	return &components.ContainerColumn{
		Page: components.Page{Key: "seer_reddit.RedditSourceCreateFormFields"},
		Children: []components.PageInterface{
			&components.ContainerError{
				Page:  components.Page{Key: "seer_reddit.RedditSourceForm.RedditRunnerID"},
				Error: getters.Key[error]("$error.RedditRunnerID"),
				Children: []components.PageInterface{
					&components.InputForeignKey[RedditRunner]{
						Label:       "Worker",
						Name:        "RedditRunnerID",
						Url:         lago.RoutePath("seer_reddit.RedditRunnerSelectRoute", nil),
						Display:     getters.Key[string]("$in.Name"),
						Placeholder: "Optional worker…",
						Required:    false,
						Getter:      getters.Association[RedditRunner](getters.Deref(getters.Key[*uint]("$in.RedditRunnerID"))),
						Classes:     "w-full max-w-xl",
					},
				},
			},
			&components.ContainerError{
				Error: getters.Key[error]("$error.Subreddits"),
				Children: []components.PageInterface{
					&InputSubredditList{
						InputStringList: components.InputStringList{
							Label:   "Subreddits",
							Name:    "Subreddits",
							Classes: "w-full max-w-xl",
							Getter:  getters.Key[datatypes.JSON]("$in.Subreddits"),
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
						Getter:  getters.Key[string]("$in.SearchQuery"),
						Classes: "w-full max-w-xl",
					},
				},
			},
			&components.ClientData{
				Page: components.Page{Key: "seer_reddit.RedditSourceForm.FilterBlock"},
				Data: "{ isFilterWhitelist: false }",
				Init: "isFilterWhitelist = $el.querySelector('[name=IsFilterWhitelist]')?.checked ?? false",
				Children: []components.PageInterface{
					&components.ContainerError{
						Error: getters.Key[error]("$error.IsFilterWhitelist"),
						Children: []components.PageInterface{
							&components.InputCheckbox{
								Page:    components.Page{Key: "seer_reddit.RedditSourceForm.IsFilterWhitelist"},
								Label:   "Treat filter as whitelist (off = blacklist)",
								Name:    "IsFilterWhitelist",
								Getter:  getters.Key[bool]("$in.IsFilterWhitelist"),
								XModel:  "isFilterWhitelist",
								Classes: "w-full max-w-xl",
							},
						},
					},
					&components.ContainerError{
						Error: getters.Key[error]("$error.Filter"),
						Children: []components.PageInterface{
							&components.ClientIf{
								Condition: "isFilterWhitelist",
								Children: []components.PageInterface{
									&components.InputTextarea{
										Label:   "Filter (whitelist: keep posts that match one of these lines)",
										Name:    "Filter",
										Rows:    4,
										Getter:  getters.Key[string]("$in.Filter"),
										Classes: "w-full max-w-xl",
									},
								},
							},
							&components.ClientIf{
								Condition: "!isFilterWhitelist",
								Children: []components.PageInterface{
									&components.InputTextarea{
										Label:   "Filter (blacklist: drop posts that match one of these lines)",
										Name:    "Filter",
										Rows:    4,
										Getter:  getters.Key[string]("$in.Filter"),
										Classes: "w-full max-w-xl",
									},
								},
							},
						},
					},
				},
			},
			&components.ContainerError{
				Error: getters.Key[error]("$error.MaxFreshPosts"),
				Children: []components.PageInterface{
					&components.InputNumber[uint]{
						Label: "Max fresh posts per fetch",
						Name:  "MaxFreshPosts",
						Getter: getters.Map(getters.Key[uint]("$in.MaxFreshPosts"), func(_ context.Context, n uint) (uint, error) {
							if n == 0 {
								return defaultMaxFreshPosts, nil
							}
							return n, nil
						}),
						Classes: "w-full max-w-md",
					},
				},
			},
			&components.ContainerError{
				Error: getters.Key[error]("$error.LoadWebsites"),
				Children: []components.PageInterface{
					&components.InputCheckbox{
						Page:    components.Page{Key: "seer_reddit.RedditSourceForm.LoadWebsites"},
						Label:   "Queue external URLs for website scrape (Seer websites worker)",
						Name:    "LoadWebsites",
						Getter:  getters.Key[bool]("$in.LoadWebsites"),
						Classes: "w-full max-w-xl",
					},
				},
			},
		},
	}
}

func registerRedditSourceUpdatePages() {
	updateFormName := getters.Static("seer_reddit.RedditSourceUpdateForm")
	deleteFormName := getters.Static("seer_reddit.RedditSourceDeleteForm")

	lago.RegistryPage.Register("seer_reddit.RedditSourceUpdateForm", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "seer_reddit.RedditSourceDetailMenu"},
		},
		Children: []components.PageInterface{
			&components.FormListenBoostedPost{
				Name: updateFormName,
				ActionURL: lago.RoutePath("seer_reddit.RedditSourceUpdateRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("redditSource.ID")),
				}),
				Children: []components.PageInterface{
					&components.FormComponent[RedditSource]{
						Getter:   getters.Key[RedditSource]("redditSource"),
						Attr:     getters.FormBubbling(updateFormName),
						Title:    "Edit Reddit source",
						Subtitle: "Subreddit names without r/ prefix. Optional search query narrows listing or search results.",
						Classes:  "@container",
						ChildrenInput: []components.PageInterface{
							redditSourceCreateFormFields(),
						},
						ChildrenAction: []components.PageInterface{
							&components.ContainerRow{
								Classes: "flex flex-wrap justify-between gap-2 mt-2 items-center",
								Children: []components.PageInterface{
									&components.ContainerRow{
										Classes: "flex justify-end gap-2",
										Children: []components.PageInterface{
											&components.ButtonSubmit{Label: "Save changes"},
											&components.ButtonModalForm{
												Label:       "Delete",
												Icon:        "trash",
												Name:        deleteFormName,
												Url:         lago.RoutePath("seer_reddit.RedditSourceDeleteRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("redditSource.ID"))}),
												FormPostURL: lago.RoutePath("seer_reddit.RedditSourceDeleteRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("redditSource.ID"))}),
												ModalUID:    "seer-reddit-source-delete-modal",
												Classes:     "btn-error",
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

	lago.RegistryPage.Register("seer_reddit.RedditSourceDeleteForm", &components.Modal{
		UID: "seer-reddit-source-delete-modal",
		Children: []components.PageInterface{
			&components.DeleteConfirmation{
				Title:   "Delete Reddit source?",
				Message: "Removes this source and its links to posts. Reddit post rows are kept; re-add a source to attach them again.",
				Attr:    getters.FormBubbling(deleteFormName),
			},
		},
	})
}

func registerRedditSourceCreatePages() {
	createFormName := getters.Static("seer_reddit.RedditSourceCreateForm")

	lago.RegistryPage.Register("seer_reddit.RedditSourceCreateForm", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "seer_reddit.RedditMenu"},
		},
		Children: []components.PageInterface{
			&components.FormListenBoostedPost{
				Name:      createFormName,
				ActionURL: lago.RoutePath("seer_reddit.RedditSourceCreateRoute", nil),
				Children: []components.PageInterface{
					&components.FormComponent[RedditSource]{
						Getter:   getters.Static(redditSourceCreateFormDefaults),
						Attr:     getters.FormBubbling(createFormName),
						Title:    "Create Reddit source",
						Subtitle: "Subreddit names without r/ prefix. Optional search query narrows listing or search results.",
						Classes:  "@container",
						ChildrenInput: []components.PageInterface{
							redditSourceCreateFormFields(),
						},
						ChildrenAction: []components.PageInterface{
							&components.ButtonSubmit{Label: "Save source"},
						},
					},
				},
			},
		},
	})
}
