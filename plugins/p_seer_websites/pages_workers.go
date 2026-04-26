package p_seer_websites

import (
	"context"
	"time"

	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
)

func websiteRunnerFormFields() components.PageInterface {
	return &components.ContainerColumn{
		Page: components.Page{Key: "seer_websites.WebsiteRunnerFormFields"},
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
		},
	}
}

func registerWebsiteRunnerPages() {
	createName := getters.Static("seer_websites.WebsiteRunnerCreateForm")
	updateName := getters.Static("seer_websites.WebsiteRunnerUpdateForm")
	deleteName := getters.Static("seer_websites.WebsiteRunnerDeleteForm")

	lago.RegistryPage.Register("seer_websites.WebsiteRunnerTable", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "seer_websites.WebsiteMenu"},
		},
		Children: []components.PageInterface{
			&components.DataTable[WebsiteRunner]{
				Page:    components.Page{Key: "seer_websites.WebsiteRunnerTableBody"},
				UID:     "seer-website-runners-table",
				Classes: "w-full",
				Data:    getters.Key[components.ObjectList[WebsiteRunner]]("websiteRunners"),
				Actions: []components.PageInterface{
					&components.TableButtonCreate{Link: lago.RoutePath("seer_websites.WebsiteRunnerCreateRoute", nil)},
				},
				RowAttr: getters.RowAttrNavigate(
					lago.RoutePath("seer_websites.WebsiteRunnerDetailRoute", map[string]getters.Getter[any]{
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

	lago.RegistryPage.Register("seer_websites.WebsiteRunnerSelectionTable", &components.Modal{
		UID: "website-runner-selection-modal",
		Children: []components.PageInterface{
			&components.DataTable[WebsiteRunner]{
				Page:    components.Page{Key: "seer_websites.WebsiteRunnerSelectionTableBody"},
				UID:     "website-runner-selection-table",
				Title:   "Select worker",
				Data:    getters.Key[components.ObjectList[WebsiteRunner]]("websiteRunners"),
				RowAttr: getters.RowAttrSelect("WebsiteRunnerID", getters.Key[uint]("$row.ID"), getters.Key[string]("$row.Name")),
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

	lago.RegistryPage.Register("seer_websites.WebsiteRunnerDetail", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "seer_websites.WebsiteRunnerDetailMenu"},
		},
		Children: []components.PageInterface{
			&components.Detail[WebsiteRunner]{
				Getter: getters.Key[WebsiteRunner]("websiteRunner"),
				Children: []components.PageInterface{
					components.ContainerColumn{
						Page: components.Page{Key: "seer_websites.WebsiteRunnerDetailContent"},
						Children: []components.PageInterface{
							&components.FieldTitle{Getter: getters.Key[string]("$in.Name")},
							&components.LabelInline{
								Title: "Duration",
								Children: []components.PageInterface{
									&components.FieldDuration{
										Getter: getters.Ref(getters.Key[time.Duration]("websiteRunner.Duration")),
									},
								},
							},
							&components.ShowIf{
								Page:   components.Page{Key: "seer_websites.WebsiteRunnerDetailWorkerPoolStopWrap"},
								Getter: getters.Any(getters.Key[bool]("workerPoolIsRunning")),
								Children: []components.PageInterface{
									&components.ContainerRow{
										Page:    components.Page{Key: "seer_websites.WebsiteRunnerDetailWorkerPoolActions"},
										Classes: "flex flex-wrap gap-2 items-center mt-2",
										Children: []components.PageInterface{
											&components.ButtonPost{
												Label: "Stop worker pool",
												URL: lago.RoutePath("seer_websites.WebsiteRunnerWorkerPoolStopRoute", map[string]getters.Getter[any]{
													"id": getters.Any(getters.Key[uint]("$in.ID")),
												}),
												Icon:    "stop",
												Classes: "btn-outline btn-error btn-sm",
											},
										},
									},
								},
							},
							&components.ShowIf{
								Page: components.Page{Key: "seer_websites.WebsiteRunnerDetailWorkerPoolStartWrap"},
								Getter: getters.Map(getters.Key[bool]("workerPoolIsRunning"), func(_ context.Context, running bool) (any, error) {
									return !running, nil
								}),
								Children: []components.PageInterface{
									&components.ContainerRow{
										Page:    components.Page{Key: "seer_websites.WebsiteRunnerDetailWorkerPoolActions"},
										Classes: "flex flex-wrap gap-2 items-center mt-2",
										Children: []components.PageInterface{
											&components.ButtonPost{
												Label: "Start worker pool",
												URL: lago.RoutePath("seer_websites.WebsiteRunnerWorkerPoolStartRoute", map[string]getters.Getter[any]{
													"id": getters.Any(getters.Key[uint]("$in.ID")),
												}),
												Icon:    "play",
												Classes: "btn-outline btn-success btn-sm",
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

	lago.RegistryPage.Register("seer_websites.WebsiteRunnerCreateForm", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "seer_websites.WebsiteMenu"},
		},
		Children: []components.PageInterface{
			&components.FormListenBoostedPost{
				Name:      createName,
				ActionURL: lago.RoutePath("seer_websites.WebsiteRunnerCreateRoute", nil),
				Children: []components.PageInterface{
					&components.FormComponent[WebsiteRunner]{
						Getter:   getters.Static(WebsiteRunner{Name: "", Duration: time.Hour}),
						Attr:     getters.FormBubbling(createName),
						Title:    "Create worker",
						Subtitle: "Workers define a name and a cadence (e.g. 5m, 1h). Website sources assigned to this worker run a recursive crawl on each tick.",
						Classes:  "@container",
						ChildrenInput: []components.PageInterface{
							websiteRunnerFormFields(),
						},
						ChildrenAction: []components.PageInterface{
							&components.ButtonSubmit{Label: "Save worker"},
						},
					},
				},
			},
		},
	})

	lago.RegistryPage.Register("seer_websites.WebsiteRunnerUpdateForm", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "seer_websites.WebsiteRunnerDetailMenu"},
		},
		Children: []components.PageInterface{
			&components.FormListenBoostedPost{
				Name: updateName,
				ActionURL: lago.RoutePath("seer_websites.WebsiteRunnerUpdateRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("websiteRunner.ID")),
				}),
				Children: []components.PageInterface{
					&components.FormComponent[WebsiteRunner]{
						Getter:   getters.Key[WebsiteRunner]("websiteRunner"),
						Attr:     getters.FormBubbling(updateName),
						Title:    "Edit worker",
						Subtitle: "Go duration syntax: 30s, 5m, 1h30m.",
						Classes:  "@container",
						ChildrenInput: []components.PageInterface{
							websiteRunnerFormFields(),
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
												Url:         lago.RoutePath("seer_websites.WebsiteRunnerDeleteRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("websiteRunner.ID"))}),
												FormPostURL: lago.RoutePath("seer_websites.WebsiteRunnerDeleteRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("websiteRunner.ID"))}),
												ModalUID:    "seer-website-runner-delete-modal",
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

	lago.RegistryPage.Register("seer_websites.WebsiteRunnerDeleteForm", &components.Modal{
		UID: "seer-website-runner-delete-modal",
		Children: []components.PageInterface{
			&components.DeleteConfirmation{
				Title:   "Delete worker?",
				Message: "Cannot delete if a website source still references this worker (database will reject).",
				Attr:    getters.FormBubbling(deleteName),
			},
		},
	})
}
