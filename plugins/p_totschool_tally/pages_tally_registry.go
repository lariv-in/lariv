package p_totschool_tally

import (
	"time"

	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_users"
)

func registerTallyPages() {
	lago.RegistryPage.Register("tally.TallyMenu", components.SidebarMenu{
		Title: getters.Static("Totschool Tally"),
		Back: &components.SidebarMenuItem{
			Title: getters.Static("Back to Home"),
			Url:   lago.RoutePath("dashboard.AppsPage", nil),
		},
		Children: []components.PageInterface{
			components.SidebarMenuItem{
				Title: getters.Static("Dashboard"),
				Url:   lago.RoutePath("tally.TallyDashboardRoute", nil),
				Icon:  "home",
			},
			components.SidebarMenuItem{
				Title: getters.Static("Leaderboard"),
				Url:   lago.RoutePath("tally.TallyLeaderboardRoute", nil),
				Icon:  "trophy",
			},
			components.SidebarMenuItem{
				Title: getters.Static("List"),
				Url:   lago.RoutePath("tally.TallyListRoute", nil),
				Icon:  "list-bullet",
			},
			components.SidebarMenuItem{
				Title: getters.Static("Fill Daily Report"),
				Url:   lago.RoutePath("tally.TallyDailyFormRoute", nil),
				Icon:  "pencil-square",
			},
			components.SidebarMenuItem{
				Page:  components.Page{Roles: []string{"totschool_admin", "superuser"}},
				Title: getters.Static("Create Tally (Admin)"),
				Url:   lago.RoutePath("tally.TallyCreateRoute", nil),
				Icon:  "plus",
			},
		},
	})

	lago.RegistryPage.Register("tally.TallyDetailMenu", components.SidebarMenu{
		Title: getters.Static("Tally Details"),
		Back: &components.SidebarMenuItem{
			// Show the user's name and the tally date (date only), using a
			// formatted time.Time getter for the Date field.
			Title: getters.Format(
				"Tally: %s (%s)",
				getters.Any(getters.Key[string]("Tally.User.Name")),
				getters.Any(getters.TimeFormat("2006-01-02", getters.Key[time.Time]("Tally.Date"))),
			),
			Url: lago.RoutePath("tally.TallyListRoute", nil),
		},
		Children: []components.PageInterface{
			components.SidebarMenuItem{
				Title: getters.Static("Details"),
				Url:   lago.RoutePath("tally.TallyDetailRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("Tally.ID"))}),
			},
			components.SidebarMenuItem{
				Title: getters.Static("Edit"),
				Url:   lago.RoutePath("tally.TallyUpdateRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("Tally.ID"))}),
			},
		},
	})

	// Daily Create Form
	lago.RegistryPage.Register("tally.TallyDailyForm", components.ShellScaffold{
		Sidebar: []components.PageInterface{lago.DynamicPage{Name: "tally.TallyMenu"}},
		Children: []components.PageInterface{
			&components.FormListenBoostedPost{
				Name:      getters.Static("tally.TallyDailyForm"),
				ActionURL: lago.RoutePath("tally.TallyDailyFormRoute", nil),
				Children: []components.PageInterface{
					components.FormComponent[Tally]{
						Attr: getters.FormBubbling(getters.Static("tally.TallyDailyForm")),

						Title:         "Daily Tally",
						Subtitle:      "Submit or update your tally for today",
						ChildrenInput: tallyCommonFields(),
						ChildrenAction: []components.PageInterface{
							components.ButtonSubmit{Label: "Submit Daily Tally"},
						},
					},
				},
			},
		},
	})

	// Create Form (Admin)
	createAdminFields := append([]components.PageInterface{
		components.InputForeignKey[p_users.User]{
			Page:        components.Page{Roles: []string{"totschool_admin", "superuser"}},
			Name:        "UserID",
			Label:       "User",
			Url:         lago.RoutePath("users.SelectRoute", nil),
			Display:     getters.Key[string]("$in.Name"),
			Placeholder: "Select a user...",
			Required:    true,
			Getter:      getters.Association[p_users.User](getters.Key[uint]("$in.UserID")),
		},
		components.InputDate{
			Page:     components.Page{Roles: []string{"totschool_admin", "superuser"}},
			Name:     "Date",
			Label:    "Date (YYYY-MM-DD)",
			Required: true,
			Getter:   getters.Key[time.Time]("$in.Date"),
		},
	}, tallyCommonFields()...)

	lago.RegistryPage.Register("tally.TallyCreateForm", components.ShellScaffold{
		Sidebar: []components.PageInterface{lago.DynamicPage{Name: "tally.TallyMenu"}},
		Children: []components.PageInterface{
			&components.FormListenBoostedPost{
				Name:      getters.Static("tally.TallyCreateForm"),
				ActionURL: lago.RoutePath("tally.TallyCreateRoute", nil),
				Children: []components.PageInterface{
					components.FormComponent[Tally]{
						Attr: getters.FormBubbling(getters.Static("tally.TallyCreateForm")),

						Title:         "Create Tally",
						Subtitle:      "Create a tally record for a specific user and date",
						ChildrenInput: createAdminFields,
						ChildrenAction: []components.PageInterface{
							components.ButtonSubmit{Label: "Save Tally"},
						},
					},
				},
			},
		},
	})

	// Update Form (Admin)
	lago.RegistryPage.Register("tally.TallyUpdateForm", components.ShellScaffold{
		Sidebar: []components.PageInterface{lago.DynamicPage{Name: "tally.TallyDetailMenu"}},
		Children: []components.PageInterface{
			&components.FormListenBoostedPost{
				Name:      getters.Static("tally.TallyUpdateForm"),
				ActionURL: lago.RoutePath("tally.TallyUpdateRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("Tally.ID"))}),
				Children: []components.PageInterface{
					components.FormComponent[Tally]{
						Attr: getters.FormBubbling(getters.Static("tally.TallyUpdateForm")),

						Title:         "Update Tally",
						Subtitle:      "Edit tally details",
						ChildrenInput: createAdminFields,
						ChildrenAction: []components.PageInterface{
							components.ContainerRow{
								Classes: "flex flex-wrap justify-between gap-2 mt-2 items-center",
								Children: []components.PageInterface{
									components.ContainerRow{
										Classes: "flex justify-end gap-2",
										Children: []components.PageInterface{
											components.ButtonSubmit{Label: "Update Tally"},
											components.ButtonModalForm{
												Page:        components.Page{Roles: []string{"totschool_admin", "superuser"}},
												Label:       "Delete",
												Icon:        "trash",
												Name:        getters.Static("tally.TallyDeleteForm"),
												Url:         lago.RoutePath("tally.TallyDeleteRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("Tally.ID"))}),
												FormPostURL: lago.RoutePath("tally.TallyDeleteRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("Tally.ID"))}),
												ModalUID:    "tally-delete-modal",
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

	// Delete Form
	lago.RegistryPage.Register("tally.TallyDeleteForm", components.Modal{
		UID: "tally-delete-modal",
		Children: []components.PageInterface{
			components.DeleteConfirmation{
				Title:   "Delete Tally?",
				Message: "Are you sure you want to delete this tally? This action cannot be undone.",
				Attr:    getters.FormBubbling(getters.Key[string]("$get.name")),
			},
		},
	})

	// Tally Detail View
	lago.RegistryPage.Register("tally.TallyDetail", components.ShellScaffold{
		Sidebar: []components.PageInterface{lago.DynamicPage{Name: "tally.TallyDetailMenu"}},
		Children: []components.PageInterface{
			components.ContainerColumn{
				Classes: "p-4",
				Children: []components.PageInterface{
					components.FieldTitle{Getter: getters.Static("Tally Details")},
				},
			},
			components.Detail[Tally]{
				Getter: getters.Key[Tally]("Tally"),
				Children: []components.PageInterface{
					components.ContainerRow{
						Classes: "grid grid-cols-1 md:grid-cols-2 gap-y-4 gap-x-8 p-4 bg-base-100 shadow rounded-box",
						Children: []components.PageInterface{
							components.ContainerColumn{
								Children: []components.PageInterface{
									components.FieldTitle{Getter: getters.Static("User")},
									components.FieldText{
										Getter:  getters.ForeignKey[p_users.User, uint, string](getters.Key[uint]("$in.UserID"), "Name"),
										Classes: "font-semibold",
									},
								},
							},
							components.ContainerColumn{
								Children: []components.PageInterface{
									components.FieldTitle{Getter: getters.Static("Date")},
									components.FieldDatetime{Getter: getters.Key[time.Time]("$in.Date"), Classes: "font-semibold"},
								},
							},
							components.ContainerColumn{
								Children: []components.PageInterface{
									components.FieldTitle{Getter: getters.Static("Visits")},
									components.FieldText{
										Getter:  getters.IntString(getters.Key[int]("$in.Visits")),
										Classes: "font-semibold",
									},
								},
							},
							components.ContainerColumn{
								Children: []components.PageInterface{
									components.FieldTitle{Getter: getters.Static("Appointments")},
									components.FieldText{
										Getter:  getters.IntString(getters.Key[int]("$in.Appointments")),
										Classes: "font-semibold",
									},
								},
							},
							components.ContainerColumn{
								Children: []components.PageInterface{
									components.FieldTitle{Getter: getters.Static("Leads")},
									components.FieldText{
										Getter:  getters.IntString(getters.Key[int]("$in.Leads")),
										Classes: "font-semibold",
									},
								},
							},
							components.ContainerColumn{
								Children: []components.PageInterface{
									components.FieldTitle{Getter: getters.Static("Presentations")},
									components.FieldText{
										Getter:  getters.IntString(getters.Key[int]("$in.Presentations")),
										Classes: "font-semibold",
									},
								},
							},
							components.ContainerColumn{
								Children: []components.PageInterface{
									components.FieldTitle{Getter: getters.Static("Demonstrations")},
									components.FieldText{
										Getter:  getters.IntString(getters.Key[int]("$in.Demos")),
										Classes: "font-semibold",
									},
								},
							},
							components.ContainerColumn{
								Children: []components.PageInterface{
									components.FieldTitle{Getter: getters.Static("Follow Up Letters Sent")},
									components.FieldText{
										Getter:  getters.IntString(getters.Key[int]("$in.Letters")),
										Classes: "font-semibold",
									},
								},
							},
							components.ContainerColumn{
								Children: []components.PageInterface{
									components.FieldTitle{Getter: getters.Static("Follow Ups")},
									components.FieldText{
										Getter:  getters.IntString(getters.Key[int]("$in.FollowUps")),
										Classes: "font-semibold",
									},
								},
							},
							components.ContainerColumn{
								Children: []components.PageInterface{
									components.FieldTitle{Getter: getters.Static("Proposals Given")},
									components.FieldText{
										Getter:  getters.IntString(getters.Key[int]("$in.Proposals")),
										Classes: "font-semibold",
									},
								},
							},
							components.ContainerColumn{
								Children: []components.PageInterface{
									components.FieldTitle{Getter: getters.Static("Policies Sold")},
									components.FieldText{
										Getter:  getters.IntString(getters.Key[int]("$in.Policies")),
										Classes: "font-semibold",
									},
								},
							},
							components.ContainerColumn{
								Children: []components.PageInterface{
									components.FieldTitle{Getter: getters.Static("Premium")},
									components.FieldText{
										Getter:  getters.IntString(getters.Key[int]("$in.Premium")),
										Classes: "font-semibold",
									},
								},
							},
						},
					},
				},
			},
		},
	})

	// Tally Filter
	tallyFilter := components.FormComponent[Tally]{
		Attr: getters.FormBoostedGet(lago.RoutePath("tally.TallyListRoute", nil)),

		ChildrenInput: []components.PageInterface{
			components.InputForeignKey[uint]{
				Page: components.Page{Roles: []string{"totschool_admin", "superuser"}},

				Name:    "UserID",
				Label:   "User ID",
				Url:     lago.RoutePath("users.SelectRoute", nil),
				Getter:  getters.Key[uint]("$get.UserID"),
				Display: getters.Key[string]("$in.Name"),
			},
			components.InputDate{Name: "Date", Label: "Date", Getter: getters.Key[time.Time]("$get.Date")},
		},
		ChildrenAction: []components.PageInterface{
			components.ButtonSubmit{Label: "Apply Filter"},
			components.ButtonClear{Label: "Clear"},
		},
	}

	// Session environment selector (shared across list, dashboard, leaderboard)
	sessionEnvironment := &components.Environment[uint]{
		Label:   "Session",
		Key:     getters.Static("session"),
		Options: SessionsListGetter,
		Default: tallySessionEnvironmentDefault,
	}

	// Tally Table
	lago.RegistryPage.Register("tally.TallyTable", components.ShellScaffold{
		Sidebar: []components.PageInterface{lago.DynamicPage{Name: "tally.TallyMenu"}},
		Children: []components.PageInterface{
			sessionEnvironment,
			components.DataTable[Tally]{
				Title:    "Tallies List",
				Subtitle: "All tallies in the system",
				Data:     getters.Key[components.ObjectList[Tally]]("Tallies"),
				Actions: []components.PageInterface{
					&components.TableButtonFilter{Child: &tallyFilter},
				},
				Classes: "mt-4",
				Columns: []components.TableColumn{
					{
						Label: "Date",
						Name:  "Date",
						Children: []components.PageInterface{
							components.FieldDatetime{Getter: getters.Key[time.Time]("$row.Date")},
						},
					},
					{
						Label: "User",
						Name:  "User.Name",
						Children: []components.PageInterface{
							components.FieldText{
								Getter: getters.Key[string]("$row.User.Name"),
							},
						},
					},
					{
						Label: "Visits",
						Name:  "Visits",
						Children: []components.PageInterface{
							components.FieldText{
								Getter: getters.IntString(getters.Key[int]("$row.Visits")),
							},
						},
					},
					{
						Label: "Appointments",
						Name:  "Appointments",
						Children: []components.PageInterface{
							components.FieldText{
								Getter: getters.IntString(getters.Key[int]("$row.Appointments")),
							},
						},
					},
					{
						Label: "Policies",
						Name:  "Policies",
						Children: []components.PageInterface{
							components.FieldText{
								Getter: getters.IntString(getters.Key[int]("$row.Policies")),
							},
						},
					},
					{
						Label: "Premium",
						Name:  "Premium",
						Children: []components.PageInterface{
							components.FieldText{
								Getter: getters.IntString(getters.Key[int]("$row.Premium")),
							},
						},
					},
				},
				RowAttr: getters.RowAttrNavigateFormat("/tally/%v/", getters.Any(getters.Key[uint]("$row.ID"))),
			},
		},
	})

	// Dashboard and Leaderboard rendering in pages requires a custom component or HTML container.
	lago.RegistryPage.Register("tally.TallyDashboard", components.ShellScaffold{
		Sidebar: []components.PageInterface{lago.DynamicPage{Name: "tally.TallyMenu"}},
		Children: []components.PageInterface{
			sessionEnvironment,
			components.ContainerHTML{
				HTML: TallyDashboardHTML,
			},
		},
	})

	lago.RegistryPage.Register("tally.TallyLeaderboard", components.ShellScaffold{
		Sidebar: []components.PageInterface{lago.DynamicPage{Name: "tally.TallyMenu"}},
		Children: []components.PageInterface{
			sessionEnvironment,
			components.ContainerHTML{
				HTML: TallyLeaderboardHTML,
			},
		},
	})
}
