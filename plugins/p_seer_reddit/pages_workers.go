package p_seer_reddit

import (
	"context"

	"time"

	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	"gorm.io/gorm"
)

func redditRunnerFormFields() components.PageInterface {
	return &components.ContainerColumn{
		Page: components.Page{Key: "seer_reddit.RedditRunnerFormFields"},
		Children: []components.PageInterface{
			&components.ContainerError{
				Error: getters.Key[error]("$error.Name"),
				Children: []components.PageInterface{
					&components.InputText{
						Label:    "Name",
						Name:     "Name",
						Required: true,
						Getter:   getters.Key[string]("$in.Name"),
						Classes:  "w-full max-w-xl",
					},
				},
			},
			&components.ContainerError{
				Error: getters.Key[error]("$error.Duration"),
				Children: []components.PageInterface{
					&components.InputDuration{
						Label:    "Duration",
						Name:     "Duration",
						Required: true,
						Getter:   getters.Ref(getters.Key[time.Duration]("$in.Duration")),
						Classes:  "w-full max-w-xl",
					},
				},
			},
			&components.ContainerError{
				Error: getters.Key[error]("$error.RedditSourceIDs"),
				Children: []components.PageInterface{
					&components.InputManyToMany[RedditSource]{
						Label:       "Reddit sources without worker",
						Name:        "RedditSourceIDs",
						Getter:      redditSourcesForCurrentRunner,
						Url:         lago.RoutePath("seer_reddit.RedditSourceUnsetSelectRoute", nil),
						Display:     redditSourceSelectionDisplayFromIn,
						Placeholder: "Select unassigned sources...",
						Classes:     "w-full max-w-xl",
					},
				},
			},
		},
	}
}

func redditSourcesForCurrentRunner(ctx context.Context) ([]RedditSource, error) {
	id, err := getters.Key[uint]("$in.ID")(ctx)
	if err != nil || id == 0 {
		return nil, err
	}
	db, err := getters.DBFromContext(ctx)
	if err != nil {
		return nil, err
	}
	return gorm.G[RedditSource](db).Where("reddit_runner_id = ?", id).Order("id DESC").Find(ctx)
}

func redditRunnerDetailWorkerPoolActionsGetter() getters.Getter[components.PageInterface] {
	return func(ctx context.Context) (components.PageInterface, error) {
		id, err := getters.Key[uint]("$in.ID")(ctx)
		if err != nil {
			return nil, err
		}
		if RedditRunnerWorkerPoolIsRunning(id) {
			return &components.ContainerRow{
				Page:    components.Page{Key: "seer_reddit.RedditRunnerDetailWorkerPoolActions"},
				Classes: "flex flex-wrap gap-2 items-center mt-2",
				Children: []components.PageInterface{
					&components.ButtonPost{
						Label: "Stop worker pool",
						URL: lago.RoutePath("seer_reddit.RedditRunnerWorkerPoolStopRoute", map[string]getters.Getter[any]{
							"id": getters.Any(getters.Key[uint]("$in.ID")),
						}),
						Icon:    "stop",
						Classes: "btn-outline btn-error btn-sm",
					},
				},
			}, nil
		}
		return &components.ContainerRow{
			Page:    components.Page{Key: "seer_reddit.RedditRunnerDetailWorkerPoolActions"},
			Classes: "flex flex-wrap gap-2 items-center mt-2",
			Children: []components.PageInterface{
				&components.ButtonPost{
					Label: "Start worker pool",
					URL: lago.RoutePath("seer_reddit.RedditRunnerWorkerPoolStartRoute", map[string]getters.Getter[any]{
						"id": getters.Any(getters.Key[uint]("$in.ID")),
					}),
					Icon:    "play",
					Classes: "btn-outline btn-success btn-sm",
				},
			},
		}, nil
	}
}

func registerRedditRunnerPages() {
	createName := getters.Static("seer_reddit.RedditRunnerCreateForm")
	updateName := getters.Static("seer_reddit.RedditRunnerUpdateForm")
	deleteName := getters.Static("seer_reddit.RedditRunnerDeleteForm")

	lago.RegistryPage.Register("seer_reddit.RedditRunnerTable", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "seer_reddit.RedditMenu"},
		},
		Children: []components.PageInterface{
			&components.DataTable[RedditRunner]{
				Page:    components.Page{Key: "seer_reddit.RedditRunnerTableBody"},
				UID:     "seer-reddit-runners-table",
				Classes: "w-full",
				Data:    getters.Key[components.ObjectList[RedditRunner]]("redditRunners"),
				Actions: []components.PageInterface{
					&components.TableButtonCreate{Link: lago.RoutePath("seer_reddit.RedditRunnerCreateRoute", nil)},
				},
				RowAttr: getters.RowAttrNavigate(
					lago.RoutePath("seer_reddit.RedditRunnerDetailRoute", map[string]getters.Getter[any]{
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
						Label: "Duration",
						Children: []components.PageInterface{
							&components.FieldDuration{
								Getter: getters.Ref(getters.Key[time.Duration]("$row.Duration")),
							},
						},
					},
				},
			},
		},
	})

	lago.RegistryPage.Register("seer_reddit.RedditRunnerSelectionTable", &components.Modal{
		UID: "reddit-runner-selection-modal",
		Children: []components.PageInterface{
			&components.DataTable[RedditRunner]{
				Page:    components.Page{Key: "seer_reddit.RedditRunnerSelectionTableBody"},
				UID:     "reddit-runner-selection-table",
				Title:   "Select worker",
				Data:    getters.Key[components.ObjectList[RedditRunner]]("redditRunners"),
				RowAttr: getters.RowAttrSelect("RedditRunnerID", getters.Key[uint]("$row.ID"), getters.Key[string]("$row.Name")),
				Columns: []components.TableColumn{
					{
						Label: "Name",
						Children: []components.PageInterface{
							&components.FieldText{Getter: getters.Key[string]("$row.Name")},
						},
					},
					{
						Label: "Duration",
						Children: []components.PageInterface{
							&components.FieldDuration{
								Getter: getters.Ref(getters.Key[time.Duration]("$row.Duration")),
							},
						},
					},
				},
			},
		},
	})

	lago.RegistryPage.Register("seer_reddit.RedditRunnerDetail", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "seer_reddit.RedditRunnerDetailMenu"},
		},
		Children: []components.PageInterface{
			&components.Detail[RedditRunner]{
				Getter: getters.Key[RedditRunner]("redditRunner"),
				Children: []components.PageInterface{
					components.ContainerColumn{
						Page: components.Page{Key: "seer_reddit.RedditRunnerDetailContent"},
						Children: []components.PageInterface{
							&components.FieldTitle{Getter: getters.Key[string]("$in.Name")},
							&components.LabelInline{
								Title: "Duration",
								Children: []components.PageInterface{
									&components.FieldDuration{
										Getter: getters.Ref(getters.Key[time.Duration]("redditRunner.Duration")),
									},
								},
							},
							&components.LabelNewline{
								Title: "Assigned subreddits",
								Children: []components.PageInterface{
									&components.FieldManyToMany[RedditSource]{
										Getter:  redditSourcesForCurrentRunner,
										Display: redditSourceSelectionDisplayFromIn,
										Link: lago.RoutePath("seer_reddit.RedditSourceDetailRoute", map[string]getters.Getter[any]{
											"id": getters.Any(getters.Key[uint]("$in.ID")),
										}),
										Classes: "w-full max-w-xl",
									},
								},
							},
							&components.GetterPage{Getter: redditRunnerDetailWorkerPoolActionsGetter()},
						},
					},
				},
			},
		},
	})

	lago.RegistryPage.Register("seer_reddit.RedditRunnerCreateForm", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "seer_reddit.RedditMenu"},
		},
		Children: []components.PageInterface{
			&components.FormListenBoostedPost{
				Name:      createName,
				ActionURL: lago.RoutePath("seer_reddit.RedditRunnerCreateRoute", nil),
				Children: []components.PageInterface{
					&components.FormComponent[RedditRunner]{
						Getter:   getters.Static(RedditRunner{Name: "", Duration: time.Hour}),
						Attr:     getters.FormBubbling(createName),
						Title:    "Create worker",
						Subtitle: "Workers define a name and a duration (e.g. 5m, 1h). Used by Reddit sources and posts.",
						Classes:  "@container",
						ChildrenInput: []components.PageInterface{
							redditRunnerFormFields(),
						},
						ChildrenAction: []components.PageInterface{
							&components.ButtonSubmit{Label: "Save worker"},
						},
					},
				},
			},
		},
	})

	lago.RegistryPage.Register("seer_reddit.RedditRunnerUpdateForm", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "seer_reddit.RedditRunnerDetailMenu"},
		},
		Children: []components.PageInterface{
			&components.FormListenBoostedPost{
				Name: updateName,
				ActionURL: lago.RoutePath("seer_reddit.RedditRunnerUpdateRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("redditRunner.ID")),
				}),
				Children: []components.PageInterface{
					&components.FormComponent[RedditRunner]{
						Getter:   getters.Key[RedditRunner]("redditRunner"),
						Attr:     getters.FormBubbling(updateName),
						Title:    "Edit worker",
						Subtitle: "Go duration syntax: 30s, 5m, 1h30m.",
						Classes:  "@container",
						ChildrenInput: []components.PageInterface{
							redditRunnerFormFields(),
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
												Name:        deleteName,
												Url:         lago.RoutePath("seer_reddit.RedditRunnerDeleteRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("redditRunner.ID"))}),
												FormPostURL: lago.RoutePath("seer_reddit.RedditRunnerDeleteRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("redditRunner.ID"))}),
												ModalUID:    "seer-reddit-runner-delete-modal",
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

	lago.RegistryPage.Register("seer_reddit.RedditRunnerDeleteForm", &components.Modal{
		UID: "seer-reddit-runner-delete-modal",
		Children: []components.PageInterface{
			&components.DeleteConfirmation{
				Title:   "Delete worker?",
				Message: "Cannot delete if a Reddit source still references this worker (database will reject).",
				Attr:    getters.FormBubbling(deleteName),
			},
		},
	})
}
