package p_totschool_tally

import (
	"context"
	"fmt"
	"net/url"

	"github.com/lariv-in/components"
	"github.com/lariv-in/getters"
	"github.com/lariv-in/lago"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

type TallyDashboardComponent struct {
	components.Page
}

func createStatCard(_ context.Context, title string, value string, classes string) Node {
	return Div(Class("stat rounded-box border border-base-300"),
		Div(Class("stat-title"), Text(title)),
		Div(Class(fmt.Sprintf("stat-value %s", classes)), Text(value)),
	)
}

func (d TallyDashboardComponent) Build(ctx context.Context) Node {
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

func (d TallyDashboardComponent) GetKey() string {
	return d.Key
}

func (d TallyDashboardComponent) GetRoles() []string {
	return d.Roles
}

func (d TallyDashboardComponent) GetChildren() []components.PageInterface {
	return nil
}
