package p_totschool_tally

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/lariv-in/components"
	"github.com/lariv-in/getters"
	"github.com/lariv-in/lago"
	"github.com/lariv-in/p_users"
	"gorm.io/gorm"

	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

func CurrentSessionGetter(ctx context.Context) any {
	db := ctx.Value("$db").(*gorm.DB)
	date := time.Now()
	session := EnsureSessionForDate(db, date)
	return session.Name
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
				Icon:  "plus-circle",
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
			// Use flat $in.UserID so Detail[Tally] and CRUD views can populate
			// context consistently from the Tally struct.
			Getter: getters.GetterAssociation[p_users.User]("users", getters.GetterKey[uint]("$in.UserID")),
		},
		components.InputText{
			Page:     components.Page{Roles: []string{"totschool_admin", "superuser"}},
			Name:     "Date",
			Label:    "Date (YYYY-MM-DD)",
			Required: true,
			Getter:   getters.GetterKey[string]("$in.Date"),
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
		Options: getters.GetterKey[[]string]("$in.SessionNames"),
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

// FormatCurrencyIndian formats an integer amount using the Indian numbering system,
// e.g. 1234567 -> "₹12,34,567".
func FormatCurrencyIndian(amount int) string {
	if amount == 0 {
		return "₹0"
	}

	s := fmt.Sprintf("%d", amount)
	if len(s) <= 3 {
		return "₹" + s
	}

	result := s[len(s)-3:]
	s = s[:len(s)-3]

	for len(s) > 0 {
		if len(s) <= 2 {
			result = s + "," + result
			break
		}
		result = s[len(s)-2:] + "," + result
		s = s[:len(s)-2]
	}

	return "₹" + result
}

// BuildWhatsappMessage constructs the WhatsApp report message text
// mirroring the original Django implementation.
func BuildWhatsappMessage(data WhatsappReportData) string {
	if !data.Submitted {
		return ""
	}

	dateStr := data.Date.Format("02 Jan, 2006")

	message := "TOT School Report\n"
	message += fmt.Sprintf("Date: %s\n", dateStr)
	message += fmt.Sprintf("Name: %s\n\n", data.UserName)

	// Today / QTD / Last quarter metrics
	message += fmt.Sprintf("- Visits: %d/%d/%d\n", data.Today.TotalVisits, data.QTD.TotalVisits, data.LastQuarter.TotalVisits)
	message += fmt.Sprintf("- Appointments: %d/%d/%d\n", data.Today.TotalAppointments, data.QTD.TotalAppointments, data.LastQuarter.TotalAppointments)
	message += fmt.Sprintf("- Leads: %d/%d/%d\n", data.Today.TotalLeads, data.QTD.TotalLeads, data.LastQuarter.TotalLeads)
	message += fmt.Sprintf("- Presentations: %d/%d/%d\n", data.Today.TotalPresentations, data.QTD.TotalPresentations, data.LastQuarter.TotalPresentations)
	message += fmt.Sprintf("- Demonstrations: %d/%d/%d\n", data.Today.TotalDemos, data.QTD.TotalDemos, data.LastQuarter.TotalDemos)
	message += fmt.Sprintf("- Follow Up Letters: %d/%d/%d\n", data.Today.TotalLetters, data.QTD.TotalLetters, data.LastQuarter.TotalLetters)
	message += fmt.Sprintf("- Follow Ups: %d/%d/%d\n", data.Today.TotalFollowUps, data.QTD.TotalFollowUps, data.LastQuarter.TotalFollowUps)
	message += fmt.Sprintf("- Proposals Given: %d/%d/%d\n", data.Today.TotalProposals, data.QTD.TotalProposals, data.LastQuarter.TotalProposals)

	// Premium uses Indian currency formatting
	message += fmt.Sprintf(
		"- Premium: %s/%s/%s\n",
		FormatCurrencyIndian(data.Today.TotalPremium),
		FormatCurrencyIndian(data.QTD.TotalPremium),
		FormatCurrencyIndian(data.LastQuarter.TotalPremium),
	)

	return message
}

type DashboardStats struct {
	TotalPresentations int
	TotalLeads         int
	TotalVisits        int
	TotalAppointments  int
	TotalDemos         int
	TotalLetters       int
	TotalFollowUps     int
	TotalProposals     int
	TotalPolicies      int
	TotalPremium       int
	FormsFilled        int
	ApptVisitRatio     float64
	DemoApptRatio      float64
	PolicyDemoRatio    float64
}

type WhatsappReportData struct {
	Submitted   bool
	Today       DashboardStats
	QTD         DashboardStats
	LastQuarter DashboardStats
	UserName    string
	Date        time.Time
}

func createStatCard(_ context.Context, title string, value string, classes string) Node {
	return Div(Class("stat rounded-box border border-base-300"),
		Div(Class("stat-title"), Text(title)),
		Div(Class(fmt.Sprintf("stat-value %s", classes)), Text(value)),
	)
}

func TallyDashboardHTML(ctx context.Context, _ Node) Node {
	inMap, ok := ctx.Value("$in").(map[string]any)
	if !ok {
		return Div(Text("Error loading dashboard data"))
	}

	dashboard, ok := inMap["Dashboard"].(DashboardStats)
	if !ok {
		return Div(Text("Error parsing dashboard stats"))
	}

	// Optional WhatsApp report data (only for non-admin users).
	var whatsappSection Node
	if report, ok := inMap["WhatsappReport"].(WhatsappReportData); ok {
		if report.Submitted {
			message := BuildWhatsappMessage(report)
			encoded := url.QueryEscape(message)
			whatsappURL := fmt.Sprintf("https://wa.me/?text=%s", encoded)

			whatsappSection = Div(Class("bg-base-200 rounded-box border border-base-300 p-4 my-4"),
				H3(Class("font-bold text-lg text-base-content mb-2"), Text("Today's Report Submitted!")),
				Textarea(
					Class("textarea textarea-bordered w-full h-[15rem] font-mono text-sm shadow-inner whitespace-pre overflow-y-auto mb-2"),
					Attr("readonly", "true"),
					Text(message),
				),
				A(
					Class("btn btn-sm btn-success text-white"),
					Attr("href", whatsappURL),
					Attr("target", "_blank"),
					Text("Share on WhatsApp"),
				),
			)
		} else {
			// Daily report not submitted state
			dailyURL, _ := getters.IfOrGetter(lago.GetterRoutePath("tally.TallyDailyFormRoute", nil), ctx, "")

			whatsappSection = Div(Class("bg-base-200 rounded-box border border-base-300 p-4 mb-4"),
				Div(Class("flex flex-col sm:flex-row justify-between items-start sm:items-center gap-4"),
					Div(
						H3(Class("font-bold text-lg"), Text("Daily Report Not Submitted")),
						P(Class("text-base-content/70"), Text("You haven't submitted your daily report for today.")),
					),
					A(
						Class("btn btn-primary"),
						Attr("href", dailyURL),
						Text("Fill Daily Report"),
					),
				),
			)
		}
	}

	statsHTML := Div(Class("grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-2 mt-2"),
		createStatCard(ctx, "Total Visits", fmt.Sprintf("%d", dashboard.TotalVisits), ""),
		createStatCard(ctx, "Total Appts", fmt.Sprintf("%d", dashboard.TotalAppointments), ""),
		createStatCard(ctx, "Total Leads", fmt.Sprintf("%d", dashboard.TotalLeads), ""),
		createStatCard(ctx, "Total Presentations", fmt.Sprintf("%d", dashboard.TotalPresentations), ""),
		createStatCard(ctx, "Total Demos", fmt.Sprintf("%d", dashboard.TotalDemos), ""),
		createStatCard(ctx, "Total Letters", fmt.Sprintf("%d", dashboard.TotalLetters), ""),
		createStatCard(ctx, "Total Follow Ups", fmt.Sprintf("%d", dashboard.TotalFollowUps), ""),
		createStatCard(ctx, "Total Proposals", fmt.Sprintf("%d", dashboard.TotalProposals), ""),
		createStatCard(ctx, "Total Policies", fmt.Sprintf("%d", dashboard.TotalPolicies), ""),
		createStatCard(ctx, "Total Premium", FormatCurrencyIndian(dashboard.TotalPremium), "text-success"),
	)

	ratiosHTML := Div(Class("grid grid-cols-1 md:grid-cols-4 gap-2 mt-2"),
		createStatCard(ctx, "Appt / Visit", fmt.Sprintf("%.1f%%", dashboard.ApptVisitRatio), ""),
		createStatCard(ctx, "Demo / Appt", fmt.Sprintf("%.1f%%", dashboard.DemoApptRatio), ""),
		createStatCard(ctx, "Policy / Demo", fmt.Sprintf("%.1f%%", dashboard.PolicyDemoRatio), ""),
		createStatCard(ctx, "Forms Filled", fmt.Sprintf("%d", dashboard.FormsFilled), ""),
	)

	return Div(
		If(whatsappSection != nil, whatsappSection),
		Div(Class("text-xl font-bold mt-4"), Text("Tally Dashboard")),
		statsHTML,
		Div(Class("text-xl font-bold mt-4"), Text("Conversion Rates")),
		ratiosHTML,
	)
}

type LeaderboardEntry struct {
	Rank     string
	UserID   uint
	UserName string
	Value    int
}

type LeaderboardResult struct {
	Top5        []LeaderboardEntry
	CurrentUser *LeaderboardEntry
}

func TallyLeaderboardHTML(ctx context.Context, _ Node) Node {
	inMap, ok := ctx.Value("$in").(map[string]any)
	if !ok {
		return Div(Text("Error loading leaderboard data"))
	}

	leaderboards, ok := inMap["Leaderboards"].(map[string]LeaderboardResult)
	if !ok {
		return Div(Text("Error parsing leaderboard stats"))
	}

	title, _ := inMap["Title"].(string)

	metrics := []string{"visits", "demos", "policies", "premium"}
	metricTitles := map[string]string{
		"visits":   "Top Visits",
		"demos":    "Top Demonstrations",
		"policies": "Top Policies",
		"premium":  "Top Premium",
	}

	boardsHTML := Group{}
	for _, metric := range metrics {
		board, exists := leaderboards[metric]
		if !exists {
			continue
		}

		rowsNodes := Group{}
		for _, entry := range board.Top5 {
			rowsNodes = append(rowsNodes, Tr(
				Td(Text(entry.Rank)),
				Td(Text(entry.UserName)),
				Td(Text(fmt.Sprintf("%d", entry.Value))),
			))
		}

		// Add current user summary row if present
		if board.CurrentUser != nil {
			rowsNodes = append(rowsNodes, Tr(Class("bg-base-200 font-bold"),
				Td(Text(board.CurrentUser.Rank)),
				Td(Text(board.CurrentUser.UserName+" (You)")),
				Td(Text(fmt.Sprintf("%d", board.CurrentUser.Value))),
			))
		}

		tableNode := Table(Class("table w-full"),
			THead(
				Tr(
					Th(Text("Rank")),
					Th(Text("Name")),
					Th(Text(strings.Title(metric))),
				),
			),
			TBody(rowsNodes),
		)

		boardsHTML = append(boardsHTML, Div(Class("card bg-base-100 border border-base-300 rounded-box"),
			Div(Class("card-body"),
				H2(Class("card-title"), Text(metricTitles[metric])),
				tableNode,
			),
		))
	}

	return Div(
		If(title != "", Div(Class("text-xl font-bold mt-4"), Text(title))),
		Div(Class("grid grid-cols-1 md:grid-cols-2 gap-2 mt-2"), boardsHTML),
	)
}
