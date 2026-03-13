package p_totschool_tally

import (
	"context"
	"fmt"

	"github.com/lariv-in/components"
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

	statsHTML := Div(Class("grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-2 mt-2"),
		createStatCard(ctx, "Total Visits", fmt.Sprintf("%d", dashboard.TotalVisits), "users"),
		createStatCard(ctx, "Total Appts", fmt.Sprintf("%d", dashboard.TotalAppointments), "calendar"),
		createStatCard(ctx, "Total Demos", fmt.Sprintf("%d", dashboard.TotalDemos), "presentation-chart-line"),
		createStatCard(ctx, "Total Policies", fmt.Sprintf("%d", dashboard.TotalPolicies), "document-check"),
		createStatCard(ctx, "Letters Sent", fmt.Sprintf("%d", dashboard.TotalLetters), "envelope"),
		createStatCard(ctx, "Proposals Given", fmt.Sprintf("%d", dashboard.TotalProposals), "document-text"),
		createStatCard(ctx, "Premium", fmt.Sprintf("₹%d", dashboard.TotalPremium), "currency-rupee"),
	)

	ratiosHTML := Div(Class("grid grid-cols-1 md:grid-cols-3 gap-2 mt-2"),
		createStatCard(ctx, "Appt / Visit", fmt.Sprintf("%.1f%%", dashboard.ApptVisitRatio), "chart-bar"),
		createStatCard(ctx, "Demo / Appt", fmt.Sprintf("%.1f%%", dashboard.DemoApptRatio), "chart-pie"),
		createStatCard(ctx, "Policy / Demo", fmt.Sprintf("%.1f%%", dashboard.PolicyDemoRatio), "arrow-trending-up"),
	)

	return Div(
		Div(Class("text-xl font-bold mt-4"), Text("Tally Dashboard")),
		statsHTML,
		Div(Class("text-xl font-bold mt-4"), Text("Conversion Rates")),
		ratiosHTML,
	)
}

func (d TallyDashboardComponent) GetChildren() []components.PageInterface {
	return nil
}
