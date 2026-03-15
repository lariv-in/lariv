package p_totschool_tally

import (
	"context"
	"fmt"
	"net/url"

	"github.com/lariv-in/components"
	"github.com/lariv-in/lago"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

type TallyDashboardComponent struct {
	components.Page
}

func createStatCard(ctx context.Context, title string, value string, icon string) Node {
	return Div(Class("stat bg-base-100 shadow rounded-box"),
		Div(Class("stat-figure text-primary"),
			components.Render(components.Icon{Name: icon}, ctx),
		),
		Div(Class("stat-title"), Text(title)),
		Div(Class("stat-value text-primary"), Text(value)),
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
			dailyAny := lago.GetterRoutePath("tally.TallyDailyFormRoute", nil)(ctx)
			dailyURL, _ := dailyAny.(string)

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
		createStatCard(ctx, "Total Visits", fmt.Sprintf("%d", dashboard.TotalVisits), "users"),
		createStatCard(ctx, "Total Appts", fmt.Sprintf("%d", dashboard.TotalAppointments), "calendar"),
		createStatCard(ctx, "Total Demos", fmt.Sprintf("%d", dashboard.TotalDemos), "presentation-chart-line"),
		createStatCard(ctx, "Total Policies", fmt.Sprintf("%d", dashboard.TotalPolicies), "document-check"),
		createStatCard(ctx, "Letters Sent", fmt.Sprintf("%d", dashboard.TotalLetters), "envelope"),
		createStatCard(ctx, "Proposals Given", fmt.Sprintf("%d", dashboard.TotalProposals), "document-text"),
		createStatCard(ctx, "Premium", FormatCurrencyIndian(dashboard.TotalPremium), "currency-rupee"),
	)

	ratiosHTML := Div(Class("grid grid-cols-1 md:grid-cols-4 gap-2 mt-2"),
		createStatCard(ctx, "Appt / Visit", fmt.Sprintf("%.1f%%", dashboard.ApptVisitRatio), "chart-bar"),
		createStatCard(ctx, "Demo / Appt", fmt.Sprintf("%.1f%%", dashboard.DemoApptRatio), "chart-pie"),
		createStatCard(ctx, "Policy / Demo", fmt.Sprintf("%.1f%%", dashboard.PolicyDemoRatio), "arrow-trending-up"),
		createStatCard(ctx, "Forms Filled", fmt.Sprintf("%d", dashboard.FormsFilled), "clipboard-list"),
	)

	return Div(
		If(whatsappSection != nil, whatsappSection),
		Div(Class("text-xl font-bold mt-4"), Text("Tally Dashboard")),
		statsHTML,
		Div(Class("text-xl font-bold mt-4"), Text("Conversion Rates")),
		ratiosHTML,
	)
}

func (d TallyDashboardComponent) GetChildren() []components.PageInterface {
	return nil
}
