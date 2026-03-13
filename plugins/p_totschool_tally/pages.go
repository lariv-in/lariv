package p_totschool_tally

import (
	"github.com/lariv-in/components"
	"github.com/lariv-in/getters"
	"github.com/lariv-in/lago"
)

func tallyCommonFields() []components.PageInterface {
	return []components.PageInterface{
		components.ContainerRow{
			Classes: "grid grid-cols-1 md:grid-cols-2 gap-4",
			Children: []components.PageInterface{
				components.InputNumber{Name: "Visits", Label: "Visits", Required: true, Getter: getters.GetterKey("$in.Tally.Visits")},
				components.InputNumber{Name: "Appointments", Label: "Appointments", Required: true, Getter: getters.GetterKey("$in.Tally.Appointments")},
				components.InputNumber{Name: "Leads", Label: "Leads", Required: true, Getter: getters.GetterKey("$in.Tally.Leads")},
				components.InputNumber{Name: "Presentations", Label: "Presentations", Required: true, Getter: getters.GetterKey("$in.Tally.Presentations")},
				components.InputNumber{Name: "Demos", Label: "Demonstrations", Required: true, Getter: getters.GetterKey("$in.Tally.Demos")},
				components.InputNumber{Name: "Letters", Label: "Follow Up Letters Sent", Required: true, Getter: getters.GetterKey("$in.Tally.Letters")},
				components.InputNumber{Name: "FollowUps", Label: "Follow Ups", Required: true, Getter: getters.GetterKey("$in.Tally.FollowUps")},
				components.InputNumber{Name: "Proposals", Label: "Proposals Given", Required: true, Getter: getters.GetterKey("$in.Tally.Proposals")},
				components.InputNumber{Name: "Policies", Label: "Policies Sold", Required: true, Getter: getters.GetterKey("$in.Tally.Policies")},
				components.InputNumber{Name: "Premium", Label: "Premium", Required: true, Getter: getters.GetterKey("$in.Tally.Premium")},
			},
		},
	}
}

func init() {
	lago.RegistryPage.Register("tally.TallyMenu", components.SidebarMenu{
		Title: getters.GetterStatic("Totschool Tally"),
		Back: &components.SidebarMenuItem{
			Title: getters.GetterStatic("Back to Home"),
			Url:   getters.GetterStatic("/apps/"),
		},
		Children: []components.PageInterface{
			components.SidebarMenuItem{
				Title: getters.GetterStatic("Dashboard"),
				Url:   getters.GetterStatic("/tally/"),
				Icon:  "home",
			},
			components.SidebarMenuItem{
				Title: getters.GetterStatic("Leaderboard"),
				Url:   getters.GetterStatic("/tally/leaderboard/"),
				Icon:  "trophy",
			},
			components.SidebarMenuItem{
				Title: getters.GetterStatic("List"),
				Url:   getters.GetterStatic("/tally/list/"),
				Icon:  "list-bullet",
			},
			components.SidebarMenuItem{
				Title: getters.GetterStatic("Fill Daily Report"),
				Url:   getters.GetterStatic("/tally/daily/"),
				Icon:  "pencil-square",
			},
			components.SidebarMenuItem{
				Page:  components.Page{RenderKeys: []string{"totschool_admin", "superuser"}},
				Title: getters.GetterStatic("Create Tally (Admin)"),
				Url:   getters.GetterStatic("/tally/create/"),
				Icon:  "plus-circle",
			},
		},
	})

	lago.RegistryPage.Register("tally.TallyDetailMenu", components.SidebarMenu{
		Title: getters.GetterStatic("Tally Details"),
		Back: &components.SidebarMenuItem{
			Title: getters.GetterStatic("Back to Tally List"),
			Url:   getters.GetterStatic("/tally/list/"),
		},
		Children: []components.PageInterface{
			components.SidebarMenuItem{
				Title: getters.GetterStatic("Details"),
				Url:   getters.GetterFormat("/tally/%v/", getters.GetterKey("$in.Tally.ID")),
			},
			components.SidebarMenuItem{
				Title: getters.GetterStatic("Edit"),
				Url:   getters.GetterFormat("/tally/%v/update/", getters.GetterKey("$in.Tally.ID")),
			},
			components.SidebarMenuItem{
				Title: getters.GetterStatic("Delete"),
				Url:   getters.GetterFormat("/tally/%v/delete/", getters.GetterKey("$in.Tally.ID")),
			},
		},
	})

	// Daily Create Form
	lago.RegistryPage.Register("tally.TallyDailyForm", components.ShellScaffold{
		Sidebar: []components.PageInterface{lago.DynamicPage{Name: "tally.TallyMenu"}},
		Children: []components.PageInterface{
			components.FormComponent{
				Url:           getters.GetterStatic("/tally/daily/"),
				Method:        "POST",
				Title:         "Daily Tally",
				Subtitle:      "Submit or update your tally for today",
				ChildrenInput: tallyCommonFields(),
				ChildrenAction: []components.PageInterface{
					components.ButtonSubmit{Label: "Submit Daily Tally"},
				},
			},
		},
	})

	// Create Form (Admin)
	createAdminFields := append([]components.PageInterface{
		components.InputNumber{
			Page:     components.Page{RenderKeys: []string{"totschool_admin", "superuser"}},
			Name:     "UserID",
			Label:    "User ID",
			Required: true,
			Getter:   getters.GetterKey("$in.Tally.UserID"),
		},
		components.InputText{
			Page:     components.Page{RenderKeys: []string{"totschool_admin", "superuser"}},
			Name:     "Date",
			Label:    "Date (YYYY-MM-DD)",
			Required: true,
			Getter:   getters.GetterKey("$in.Tally.Date"),
		},
	}, tallyCommonFields()...)

	lago.RegistryPage.Register("tally.TallyCreateForm", components.ShellScaffold{
		Sidebar: []components.PageInterface{lago.DynamicPage{Name: "tally.TallyMenu"}},
		Children: []components.PageInterface{
			components.FormComponent{
				Url:           getters.GetterStatic("/tally/create/"),
				Method:        "POST",
				Title:         "Create Tally",
				Subtitle:      "Create a tally record for a specific user and date",
				ChildrenInput: createAdminFields,
				ChildrenAction: []components.PageInterface{
					components.ButtonSubmit{Label: "Save Tally"},
				},
			},
		},
	})

	// Update Form (Admin)
	lago.RegistryPage.Register("tally.TallyUpdateForm", components.ShellScaffold{
		Sidebar: []components.PageInterface{lago.DynamicPage{Name: "tally.TallyDetailMenu"}},
		Children: []components.PageInterface{
			components.FormComponent{
				Url:           getters.GetterFormat("/tally/%v/update/", getters.GetterKey("$in.Tally.ID")),
				Method:        "POST",
				Title:         "Update Tally",
				Subtitle:      "Edit tally details",
				ChildrenInput: createAdminFields,
				ChildrenAction: []components.PageInterface{
					components.ButtonSubmit{Label: "Update Tally"},
				},
			},
		},
	})

	// Delete Form
	lago.RegistryPage.Register("tally.TallyDeleteForm", components.ShellScaffold{
		Sidebar: []components.PageInterface{lago.DynamicPage{Name: "tally.TallyDetailMenu"}},
		Children: []components.PageInterface{
			components.DeleteConfirmation{
				Title:     "Delete Tally?",
				Message:   "Are you sure you want to delete this tally? This action cannot be undone.",
				CancelUrl: getters.GetterFormat("/tally/%v/update/", getters.GetterKey("$in.Tally.ID")),
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
					components.FieldTitle{Getter: getters.GetterStatic("Tally Details")},
				},
			},
			components.Detail{
				Getter: getters.GetterKey("$in.Tally"),
				Children: []components.PageInterface{
					components.ContainerRow{
						Classes: "grid grid-cols-1 md:grid-cols-2 gap-y-4 gap-x-8 p-4 bg-base-100 shadow rounded-box",
						Children: []components.PageInterface{
							components.ContainerColumn{
								Children: []components.PageInterface{
									components.FieldTitle{Getter: getters.GetterStatic("User ID")},
									components.FieldText{Getter: getters.GetterKey("$in.UserID"), Classes: "font-semibold"},
								},
							},
							components.ContainerColumn{
								Children: []components.PageInterface{
									components.FieldTitle{Getter: getters.GetterStatic("Date")},
									components.FieldText{Getter: getters.GetterKey("$in.Date"), Classes: "font-semibold"},
								},
							},
							components.ContainerColumn{
								Children: []components.PageInterface{
									components.FieldTitle{Getter: getters.GetterStatic("Visits")},
									components.FieldText{Getter: getters.GetterKey("$in.Visits"), Classes: "font-semibold"},
								},
							},
							components.ContainerColumn{
								Children: []components.PageInterface{
									components.FieldTitle{Getter: getters.GetterStatic("Appointments")},
									components.FieldText{Getter: getters.GetterKey("$in.Appointments"), Classes: "font-semibold"},
								},
							},
							components.ContainerColumn{
								Children: []components.PageInterface{
									components.FieldTitle{Getter: getters.GetterStatic("Leads")},
									components.FieldText{Getter: getters.GetterKey("$in.Leads"), Classes: "font-semibold"},
								},
							},
							components.ContainerColumn{
								Children: []components.PageInterface{
									components.FieldTitle{Getter: getters.GetterStatic("Presentations")},
									components.FieldText{Getter: getters.GetterKey("$in.Presentations"), Classes: "font-semibold"},
								},
							},
							components.ContainerColumn{
								Children: []components.PageInterface{
									components.FieldTitle{Getter: getters.GetterStatic("Demonstrations")},
									components.FieldText{Getter: getters.GetterKey("$in.Demos"), Classes: "font-semibold"},
								},
							},
							components.ContainerColumn{
								Children: []components.PageInterface{
									components.FieldTitle{Getter: getters.GetterStatic("Follow Up Letters Sent")},
									components.FieldText{Getter: getters.GetterKey("$in.Letters"), Classes: "font-semibold"},
								},
							},
							components.ContainerColumn{
								Children: []components.PageInterface{
									components.FieldTitle{Getter: getters.GetterStatic("Follow Ups")},
									components.FieldText{Getter: getters.GetterKey("$in.FollowUps"), Classes: "font-semibold"},
								},
							},
							components.ContainerColumn{
								Children: []components.PageInterface{
									components.FieldTitle{Getter: getters.GetterStatic("Proposals Given")},
									components.FieldText{Getter: getters.GetterKey("$in.Proposals"), Classes: "font-semibold"},
								},
							},
							components.ContainerColumn{
								Children: []components.PageInterface{
									components.FieldTitle{Getter: getters.GetterStatic("Policies Sold")},
									components.FieldText{Getter: getters.GetterKey("$in.Policies"), Classes: "font-semibold"},
								},
							},
							components.ContainerColumn{
								Children: []components.PageInterface{
									components.FieldTitle{Getter: getters.GetterStatic("Premium")},
									components.FieldText{Getter: getters.GetterKey("$in.Premium"), Classes: "font-semibold"},
								},
							},
						},
					},
				},
			},
		},
	})

	// Tally Filter
	tallyFilter := components.FormComponent{
		Page:   components.Page{RenderKeys: []string{"totschool_admin", "superuser"}},
		Url:    getters.GetterStatic("/tally/list/"),
		Method: "GET",
		ChildrenInput: []components.PageInterface{
			components.InputNumber{Name: "UserID", Label: "User ID", Getter: getters.GetterKey("$get.UserID")},
			components.InputText{Name: "Date", Label: "Date", Getter: getters.GetterKey("$get.Date")},
		},
		ChildrenAction: []components.PageInterface{
			components.ButtonSubmit{Label: "Apply Filter"},
			components.ButtonLink{Label: "Clear", Link: getters.GetterStatic("/tally/list/")},
		},
	}

	// Session environment selector (shared across list, dashboard, leaderboard)
	sessionEnvironment := components.Environment{
		Label:   "Session",
		Key:     getters.GetterStatic("session"),
		Options: getters.GetterKey("$in.SessionNames"),
		Default: CurrentSessionNameForDateGetter,
	}

	// Tally Table
	lago.RegistryPage.Register("tally.TallyTable", components.ShellScaffold{
		Sidebar: []components.PageInterface{lago.DynamicPage{Name: "tally.TallyMenu"}},
		Children: []components.PageInterface{
			sessionEnvironment,
			components.DataTable{
				Title:           "Tallies List",
				Subtitle:        "All tallies in the system",
				Data:            getters.GetterKey("$in.Tallies"),
				FilterComponent: tallyFilter,
				Columns: []components.TableColumn{
					{Label: "Date", Key: "Date"},
					{Label: "User", Key: "User.Name"},
					{Label: "Visits", Key: "Visits"},
					{Label: "Appointments", Key: "Appointments"},
					{Label: "Policies", Key: "Policies"},
					{Label: "Premium", Key: "Premium"},
				},
				OnClick:   getters.GetterNavigate("/tally/%v/", getters.GetterKey("$row.ID")),
				CreateUrl: getters.GetterStatic("/tally/create/"),
			},
		},
	})

	// Dashboard and Leaderboard rendering in pages requires a custom component or HTML container.
	lago.RegistryPage.Register("tally.TallyDashboard", components.ShellScaffold{
		Sidebar: []components.PageInterface{lago.DynamicPage{Name: "tally.TallyMenu"}},
		Children: []components.PageInterface{
			sessionEnvironment,
			TallyDashboardComponent{},
		},
	})

	lago.RegistryPage.Register("tally.TallyLeaderboard", components.ShellScaffold{
		Sidebar: []components.PageInterface{lago.DynamicPage{Name: "tally.TallyMenu"}},
		Children: []components.PageInterface{
			sessionEnvironment,
			TallyLeaderboardComponent{},
		},
	})
}
