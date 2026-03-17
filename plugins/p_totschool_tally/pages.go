package p_totschool_tally

import (
	"context"
	"fmt"
	"time"

	"github.com/lariv-in/components"
	"github.com/lariv-in/getters"
	"github.com/lariv-in/lago"
	"github.com/lariv-in/p_users"
	"gorm.io/gorm"
)

func CurrentSessionGetter(ctx context.Context) any {
	db := ctx.Value("$db").(*gorm.DB)
	date := time.Now()
	session := EnsureSessionForDate(db, date)
	return session.Name
}

// SessionsListGetter returns the list of all session names for use in
// the session environment selector.
func SessionsListGetter(ctx context.Context) ([]string, error) {
	db := ctx.Value("$db").(*gorm.DB)
	names := getAllSessionNames(db)
	return names, nil
}

// CurrentEnvironmentSessionGetter resolves the active TotSchoolSession from
// the $environment cookie (or falls back to the current quarter), matching
// the behaviour used on tally dashboard/list pages.
func CurrentEnvironmentSessionGetter(ctx context.Context) (TotSchoolSession, error) {
	db, ok := ctx.Value("$db").(*gorm.DB)
	if !ok || db == nil {
		return TotSchoolSession{}, fmt.Errorf("TallySessionEntries: missing $db in context")
	}
	session := getSessionFromEnvironment(db, ctx)
	return session, nil
}

func tallyCommonFields() []components.PageInterface {
	return []components.PageInterface{
		components.ContainerRow{
			Classes: "grid grid-cols-1 md:grid-cols-2 gap-4",
			Children: []components.PageInterface{
				components.InputNumber{Name: "Visits", Label: "Visits", Required: true, Getter: getters.GetterKey[int]("$in.Visits")},
				components.InputNumber{Name: "Appointments", Label: "Appointments", Required: true, Getter: getters.GetterKey[int]("$in.Appointments")},
				components.InputNumber{Name: "Leads", Label: "Leads", Required: true, Getter: getters.GetterKey[int]("$in.Leads")},
				components.InputNumber{Name: "Presentations", Label: "Presentations", Required: true, Getter: getters.GetterKey[int]("$in.Presentations")},
				components.InputNumber{Name: "Demos", Label: "Demonstrations", Required: true, Getter: getters.GetterKey[int]("$in.Demos")},
				components.InputNumber{Name: "Letters", Label: "Follow Up Letters Sent", Required: true, Getter: getters.GetterKey[int]("$in.Letters")},
				components.InputNumber{Name: "FollowUps", Label: "Follow Ups", Required: true, Getter: getters.GetterKey[int]("$in.FollowUps")},
				components.InputNumber{Name: "Proposals", Label: "Proposals Given", Required: true, Getter: getters.GetterKey[int]("$in.Proposals")},
				components.InputNumber{Name: "Policies", Label: "Policies Sold", Required: true, Getter: getters.GetterKey[int]("$in.Policies")},
				components.InputNumber{Name: "Premium", Label: "Premium", Required: true, Getter: getters.GetterKey[int]("$in.Premium")},
			},
		},
	}
}

func init() {
	lago.RegistryPage.Register("tally.TallyMenu", components.SidebarMenu{
		Title: getters.GetterStatic("Totschool Tally"),
		Back: &components.SidebarMenuItem{
			Title: getters.GetterStatic("Back to Home"),
			Url:   lago.GetterRoutePath("dashboard.AppsPage", nil),
		},
		Children: []components.PageInterface{
			components.SidebarMenuItem{
				Title: getters.GetterStatic("Dashboard"),
				Url:   lago.GetterRoutePath("tally.TallyDashboardRoute", nil),
				Icon:  "home",
			},
			components.SidebarMenuItem{
				Title: getters.GetterStatic("Leaderboard"),
				Url:   lago.GetterRoutePath("tally.TallyLeaderboardRoute", nil),
				Icon:  "trophy",
			},
			components.SidebarMenuItem{
				Title: getters.GetterStatic("List"),
				Url:   lago.GetterRoutePath("tally.TallyListRoute", nil),
				Icon:  "list-bullet",
			},
			components.SidebarMenuItem{
				Title: getters.GetterStatic("Fill Daily Report"),
				Url:   lago.GetterRoutePath("tally.TallyDailyFormRoute", nil),
				Icon:  "pencil-square",
			},
			components.SidebarMenuItem{
				Page:  components.Page{Roles: []string{"totschool_admin", "superuser"}},
				Title: getters.GetterStatic("Create Tally (Admin)"),
				Url:   lago.GetterRoutePath("tally.TallyCreateRoute", nil),
				Icon:  "plus",
			},
		},
	})

	lago.RegistryPage.Register("tally.TallyDetailMenu", components.SidebarMenu{
		Title: getters.GetterStatic("Tally Details"),
		Back: &components.SidebarMenuItem{
			// Show the user's name and the tally date (date only), using a
			// formatted time.Time getter for the Date field.
			Title: getters.GetterFormat(
				"Tally: %s (%s)",
				getters.GetterAny(getters.GetterKey[string]("Tally.User.Name")),
				getters.GetterAny(getters.GetterTimeFormat("2006-01-02", getters.GetterKey[time.Time]("Tally.Date"))),
			),
			Url: lago.GetterRoutePath("tally.TallyListRoute", nil),
		},
		Children: []components.PageInterface{
			components.SidebarMenuItem{
				Title: getters.GetterStatic("Details"),
				Url:   lago.GetterRoutePath("tally.TallyDetailRoute", map[string]getters.Getter[any]{"id": getters.GetterAny(getters.GetterKey[uint]("Tally.ID"))}),
			},
			components.SidebarMenuItem{
				Title: getters.GetterStatic("Edit"),
				Url:   lago.GetterRoutePath("tally.TallyUpdateRoute", map[string]getters.Getter[any]{"id": getters.GetterAny(getters.GetterKey[uint]("Tally.ID"))}),
			},
			components.SidebarMenuItem{
				Title: getters.GetterStatic("Delete"),
				Url:   lago.GetterRoutePath("tally.TallyDeleteRoute", map[string]getters.Getter[any]{"id": getters.GetterAny(getters.GetterKey[uint]("Tally.ID"))}),
			},
		},
	})

	// Daily Create Form
	lago.RegistryPage.Register("tally.TallyDailyForm", components.ShellScaffold{
		Sidebar: []components.PageInterface{lago.DynamicPage{Name: "tally.TallyMenu"}},
		Children: []components.PageInterface{
			components.FormComponent[Tally]{
				Url:           lago.GetterRoutePath("tally.TallyDailyFormRoute", nil),
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
		components.InputForeignKey[p_users.User]{
			Page:        components.Page{Roles: []string{"totschool_admin", "superuser"}},
			Name:        "UserID",
			Label:       "User",
			Url:         lago.GetterRoutePath("users.SelectRoute", nil),
			Display:     getters.GetterKey[string]("$in.Name"),
			Placeholder: "Select a user...",
			Required:    true,
			Getter:      getters.GetterAssociation[p_users.User]("users", getters.GetterKey[uint]("$in.UserID")),
		},
		components.InputDate{
			Page:     components.Page{Roles: []string{"totschool_admin", "superuser"}},
			Name:     "Date",
			Label:    "Date (YYYY-MM-DD)",
			Required: true,
			Getter:   getters.GetterKey[time.Time]("$in.Date"),
		},
	}, tallyCommonFields()...)

	lago.RegistryPage.Register("tally.TallyCreateForm", components.ShellScaffold{
		Sidebar: []components.PageInterface{lago.DynamicPage{Name: "tally.TallyMenu"}},
		Children: []components.PageInterface{
			components.FormComponent[Tally]{
				Url:           lago.GetterRoutePath("tally.TallyCreateRoute", nil),
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
			components.FormComponent[Tally]{
				Url:           lago.GetterRoutePath("tally.TallyUpdateRoute", map[string]getters.Getter[any]{"id": getters.GetterAny(getters.GetterKey[uint]("$in.ID"))}),
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
				CancelUrl: lago.GetterRoutePath("tally.TallyUpdateRoute", map[string]getters.Getter[any]{"id": getters.GetterAny(getters.GetterKey[uint]("$in.ID"))}),
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
			components.Detail[Tally]{
				Getter: getters.GetterKey[Tally]("Tally"),
				Children: []components.PageInterface{
					components.ContainerRow{
						Classes: "grid grid-cols-1 md:grid-cols-2 gap-y-4 gap-x-8 p-4 bg-base-100 shadow rounded-box",
						Children: []components.PageInterface{
							components.ContainerColumn{
								Children: []components.PageInterface{
									components.FieldTitle{Getter: getters.GetterStatic("User")},
									components.FieldText{
										Getter:  getters.GetterForeignKey[p_users.User, uint, string](getters.GetterKey[uint]("$in.UserID"), "Name"),
										Classes: "font-semibold",
									},
								},
							},
							components.ContainerColumn{
								Children: []components.PageInterface{
									components.FieldTitle{Getter: getters.GetterStatic("Date")},
									components.FieldDatetime{Getter: getters.GetterKey[time.Time]("$in.Date"), Classes: "font-semibold"},
								},
							},
							components.ContainerColumn{
								Children: []components.PageInterface{
									components.FieldTitle{Getter: getters.GetterStatic("Visits")},
									components.FieldText{
										Getter:  getters.GetterIntString(getters.GetterKey[int]("$in.Visits")),
										Classes: "font-semibold",
									},
								},
							},
							components.ContainerColumn{
								Children: []components.PageInterface{
									components.FieldTitle{Getter: getters.GetterStatic("Appointments")},
									components.FieldText{
										Getter:  getters.GetterIntString(getters.GetterKey[int]("$in.Appointments")),
										Classes: "font-semibold",
									},
								},
							},
							components.ContainerColumn{
								Children: []components.PageInterface{
									components.FieldTitle{Getter: getters.GetterStatic("Leads")},
									components.FieldText{
										Getter:  getters.GetterIntString(getters.GetterKey[int]("$in.Leads")),
										Classes: "font-semibold",
									},
								},
							},
							components.ContainerColumn{
								Children: []components.PageInterface{
									components.FieldTitle{Getter: getters.GetterStatic("Presentations")},
									components.FieldText{
										Getter:  getters.GetterIntString(getters.GetterKey[int]("$in.Presentations")),
										Classes: "font-semibold",
									},
								},
							},
							components.ContainerColumn{
								Children: []components.PageInterface{
									components.FieldTitle{Getter: getters.GetterStatic("Demonstrations")},
									components.FieldText{
										Getter:  getters.GetterIntString(getters.GetterKey[int]("$in.Demos")),
										Classes: "font-semibold",
									},
								},
							},
							components.ContainerColumn{
								Children: []components.PageInterface{
									components.FieldTitle{Getter: getters.GetterStatic("Follow Up Letters Sent")},
									components.FieldText{
										Getter:  getters.GetterIntString(getters.GetterKey[int]("$in.Letters")),
										Classes: "font-semibold",
									},
								},
							},
							components.ContainerColumn{
								Children: []components.PageInterface{
									components.FieldTitle{Getter: getters.GetterStatic("Follow Ups")},
									components.FieldText{
										Getter:  getters.GetterIntString(getters.GetterKey[int]("$in.FollowUps")),
										Classes: "font-semibold",
									},
								},
							},
							components.ContainerColumn{
								Children: []components.PageInterface{
									components.FieldTitle{Getter: getters.GetterStatic("Proposals Given")},
									components.FieldText{
										Getter:  getters.GetterIntString(getters.GetterKey[int]("$in.Proposals")),
										Classes: "font-semibold",
									},
								},
							},
							components.ContainerColumn{
								Children: []components.PageInterface{
									components.FieldTitle{Getter: getters.GetterStatic("Policies Sold")},
									components.FieldText{
										Getter:  getters.GetterIntString(getters.GetterKey[int]("$in.Policies")),
										Classes: "font-semibold",
									},
								},
							},
							components.ContainerColumn{
								Children: []components.PageInterface{
									components.FieldTitle{Getter: getters.GetterStatic("Premium")},
									components.FieldText{
										Getter:  getters.GetterIntString(getters.GetterKey[int]("$in.Premium")),
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
		Url:    lago.GetterRoutePath("tally.TallyListRoute", nil),
		Method: "GET",
		ChildrenInput: []components.PageInterface{
			components.InputForeignKey[uint]{
				Page: components.Page{Roles: []string{"totschool_admin", "superuser"}},

				Name:    "UserID",
				Label:   "User ID",
				Url:     lago.GetterRoutePath("users.SelectRoute", nil),
				Getter:  getters.GetterKey[uint]("$get.UserID"),
				Display: getters.GetterKey[string]("$in.Name"),
			},
			components.InputDate{Name: "Date", Label: "Date", Getter: getters.GetterKey[time.Time]("$get.Date")},
		},
		ChildrenAction: []components.PageInterface{
			components.ButtonSubmit{Label: "Apply Filter"},
			components.ButtonClear{Label: "Clear"},
		},
	}

	// Session environment selector (shared across list, dashboard, leaderboard)
	sessionEnvironment := components.Environment{
		Label:   "Session",
		Key:     getters.GetterStatic("session"),
		Options: SessionsListGetter,
		Default: func(ctx context.Context) (string, error) {
			v := CurrentSessionGetter(ctx)
			if s, ok := v.(string); ok {
				return s, nil
			}
			return "", nil
		},
	}

	// Tally Table
	lago.RegistryPage.Register("tally.TallyTable", components.ShellScaffold{
		Sidebar: []components.PageInterface{lago.DynamicPage{Name: "tally.TallyMenu"}},
		Children: []components.PageInterface{
			sessionEnvironment,
			components.DataTable[Tally]{
				Title:           "Tallies List",
				Subtitle:        "All tallies in the system",
				Data:            getters.GetterKey[components.ObjectList[Tally]]("Tallies"),
				FilterComponent: tallyFilter,
				Classes:         "mt-4",
				Columns: []components.TableColumn{
					{
						Label: "Date",
						Key:   "Date",
						Children: []components.PageInterface{
							components.FieldDatetime{Getter: getters.GetterKey[time.Time]("$row.Date")},
						},
					},
					{
						Label: "User",
						Key:   "User.Name",
						Children: []components.PageInterface{
							components.FieldText{
								Getter: getters.GetterKey[string]("$row.User.Name"),
							},
						},
					},
					{
						Label: "Visits",
						Key:   "Visits",
						Children: []components.PageInterface{
							components.FieldText{
								Getter: getters.GetterIntString(getters.GetterKey[int]("$row.Visits")),
							},
						},
					},
					{
						Label: "Appointments",
						Key:   "Appointments",
						Children: []components.PageInterface{
							components.FieldText{
								Getter: getters.GetterIntString(getters.GetterKey[int]("$row.Appointments")),
							},
						},
					},
					{
						Label: "Policies",
						Key:   "Policies",
						Children: []components.PageInterface{
							components.FieldText{
								Getter: getters.GetterIntString(getters.GetterKey[int]("$row.Policies")),
							},
						},
					},
					{
						Label: "Premium",
						Key:   "Premium",
						Children: []components.PageInterface{
							components.FieldText{
								Getter: getters.GetterIntString(getters.GetterKey[int]("$row.Premium")),
							},
						},
					},
				},
				OnClick: getters.GetterNavigate("/tally/%v/", getters.GetterAny(getters.GetterKey[uint]("$row.ID"))),
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

func init() {
	// Patch the users.UserDetail page using InsertChildAfter to append
	// a session environment input that allows changing the active session.
	lago.RegistryPage.Patch("users.UserDetail", func(page components.PageInterface) components.PageInterface {
		if scaffold, ok := page.(*components.ShellScaffold); ok {
			// Ensure ApexCharts is loaded in the page head for StatLineChart.
			// NOTE: We originally attempted to inject ApexCharts into ExtraHead here,
			// but ContainerHTML requires a gomponents.Node signature and there is
			// currently no HTML wrapper in this package. To keep linting clean, we
			// skip injecting ExtraHead for now; StatLineChart assumes ApexCharts is
			// available globally (e.g. via the base layout).

			// Insert an environment input after the main user detail content
			// so that the session variable can be changed from this page.
			components.InsertChildAfter(scaffold,
				"users.UserDetailContent",
				func(*components.Detail[p_users.User]) components.ContainerColumn {
					return components.ContainerColumn{
						Children: []components.PageInterface{
							components.Environment{
								Label:   "Session",
								Key:     getters.GetterStatic("session"),
								Options: SessionsListGetter,
							},
							TallySessionEntries{
								Page: components.Page{
									Key: "tally.UserSessionTallies",
								},
								UserGetter:    getters.GetterKey[p_users.User]("user"),
								SessionGetter: CurrentEnvironmentSessionGetter,
							},
							StatLineChart{
								Page: components.Page{
									Key: "tally.UserSessionTalliesChart",
								},
								TalliesGetter: func(ctx context.Context) ([]Tally, error) {
									db, ok := ctx.Value("$db").(*gorm.DB)
									if !ok || db == nil {
										return nil, fmt.Errorf("StatLineChart: missing $db in context")
									}
									user, ok := ctx.Value("user").(p_users.User)
									if !ok {
										return nil, fmt.Errorf("StatLineChart: missing user in context")
									}
									session, err := CurrentEnvironmentSessionGetter(ctx)
									if err != nil {
										return nil, err
									}
									var tallies []Tally
									if err := db.
										Where("user_id = ? AND date >= ? AND date <= ?", user.ID, session.Start, session.End).
										Order("date ASC").
										Find(&tallies).Error; err != nil {
										return nil, err
									}
									return tallies, nil
								},
								Keys: []string{
									"Visits",
									"Appointments",
									"Leads",
									"Presentations",
									"Demos",
									"Letters",
									"FollowUps",
									"Proposals",
									"Policies",
									"Premium",
								},
							},
						},
					}
				},
			)

			return scaffold
		}
		panic("Base page for users.UserDetail was not ShellScaffold")
	})
}
