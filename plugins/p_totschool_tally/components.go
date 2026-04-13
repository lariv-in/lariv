package p_totschool_tally

import (
	"context"
	"fmt"
	"log/slog"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_users"

	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

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
	message += fmt.Sprintf("- Policies: %d/%d/%d\n", data.Today.TotalPolicies, data.QTD.TotalPolicies, data.LastQuarter.TotalPolicies)

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

func createStatCard(_ context.Context, title, value, classes string) Node {
	return Div(Class("stat rounded-box border border-base-300"),
		Div(Class("stat-title"), Text(title)),
		Div(Class(fmt.Sprintf("stat-value text-lg font-bold %s", classes)), Text(value)),
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
			dailyURL, _ := getters.IfOr(lago.RoutePath("tally.TallyDailyFormRoute", nil), ctx, "")

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

	statsHTML := Div(Class("grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-2 mt-2"),
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

	ratiosHTML := Div(Class("grid grid-cols-2 md:grid-cols-4 gap-2 mt-2"),
		createStatCard(ctx, "Appt / Visit", fmt.Sprintf("%.1f%%", dashboard.ApptVisitRatio), ""),
		createStatCard(ctx, "Demo / Appt", fmt.Sprintf("%.1f%%", dashboard.DemoApptRatio), ""),
		createStatCard(ctx, "Policy / Demo", fmt.Sprintf("%.1f%%", dashboard.PolicyDemoRatio), ""),
		createStatCard(ctx, "Forms Filled", fmt.Sprintf("%d", dashboard.FormsFilled), ""),
	)

	return Div(
		If(whatsappSection != nil, whatsappSection),
		Div(Class("text-xl font-bold mt-4"), Text("Analysis")),
		ratiosHTML,
		Div(Class("text-xl font-bold mt-4"), Text("Tally Dashboard")),
		statsHTML,
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

type TallySessionEntries struct {
	components.Page

	UserGetter    getters.Getter[p_users.User]
	SessionGetter getters.Getter[TotSchoolSession]
}

func (t TallySessionEntries) GetKey() string {
	return t.Key
}

func (t TallySessionEntries) GetRoles() []string {
	return t.Roles
}

func (t TallySessionEntries) GetChildren() []components.PageInterface {
	return nil
}

func (t *TallySessionEntries) SetChildren(children []components.PageInterface) {
	// no-op – this component has no children
}

func (t TallySessionEntries) Build(ctx context.Context) Node {
	db, err := getters.DBFromContext(ctx)
	if err != nil {
		slog.Error("TallySessionEntries: db from context", "error", err)
		return Div(Text("Error loading tally entries"))
	}

	if t.UserGetter == nil {
		slog.Error("TallySessionEntries: UserGetter is nil")
		return Div(Text("Error loading tally entries"))
	}
	user, err := t.UserGetter(ctx)
	if err != nil {
		slog.Error("TallySessionEntries: failed to resolve user", "error", err)
		return Div(Text("Error loading tally entries"))
	}

	if t.SessionGetter == nil {
		slog.Error("TallySessionEntries: SessionGetter is nil")
		return Div(Text("Error loading tally entries"))
	}
	session, err := t.SessionGetter(ctx)
	if err != nil {
		slog.Error("TallySessionEntries: failed to resolve session", "error", err)
		return Div(Text("Error loading tally entries"))
	}

	var tallies []Tally
	if err := db.
		Where("user_id = ? AND date >= ? AND date <= ?", user.ID, session.Start, session.End).
		Order("date ASC").
		Find(&tallies).Error; err != nil {
		slog.Error("TallySessionEntries: failed to load tallies", "error", err, "user_id", user.ID, "session", session.Name)
		return Div(Text("Error loading tally entries"))
	}

	// Build a lookup of which days in the session have tallies.
	hasTally := map[time.Time]bool{}
	for _, entry := range tallies {
		day := entry.Date.Truncate(24 * time.Hour)
		hasTally[day] = true
	}

	// Build a month x day grid for the entire session.
	type monthRow struct {
		month time.Time
		days  []time.Time
	}

	rows := []monthRow{}
	// Normalize start and end to midnight.
	start := session.Start.Truncate(24 * time.Hour)
	end := session.End.Truncate(24 * time.Hour)

	// Precompute months in the session.
	for m := time.Date(start.Year(), start.Month(), 1, 0, 0, 0, 0, start.Location()); !m.After(end); m = m.AddDate(0, 1, 0) {
		// Collect all days of this month within the session range.
		var days []time.Time
		for d := m; d.Month() == m.Month() && !d.After(end); d = d.AddDate(0, 0, 1) {
			if d.Before(start) {
				continue
			}
			days = append(days, d)
		}
		if len(days) > 0 {
			rows = append(rows, monthRow{month: m, days: days})
		}
	}

	// Render grid: rows = months, columns = days, cell color based on tally presence.
	gridRows := Group{}
	for _, row := range rows {
		dayCells := Group{}
		for _, d := range row.days {
			dayKey := d.Truncate(24 * time.Hour)
			green := hasTally[dayKey]
			cellClass := "w-6 h-6 text-[10px] flex items-center justify-center rounded"
			if green {
				cellClass += " bg-success text-success-content"
			} else {
				cellClass += " bg-error text-error-content"
			}
			dayCells = append(dayCells,
				Div(
					Class(cellClass),
					Attr("title", d.Format("2006-01-02")),
					Text(fmt.Sprintf("%d", d.Day())),
				),
			)
		}

		gridRows = append(gridRows,
			Div(
				Div(Class("mt-2"), Text(row.month.Format("Jan 2006"))),
				Div(Class("flex flex-wrap gap-1"), dayCells),
			),
		)
	}

	return Div(
		Class("mt-4"),
		Div(Class("font-semibold mb-2"),
			Text(fmt.Sprintf("Forms filled in %s", session.Name)),
		),
		Div(Class("flex flex-col gap-1"), gridRows),
	)
}

// StatLineChart is a reusable component that visualises a numeric field from
// Tally over time (per day in the selected session) using ApexCharts.
// It supports switching between multiple metric keys via Alpine.js.
type StatLineChart struct {
	components.Page

	// TalliesGetter resolves the list of tallies to chart (typically the
	// user's tallies for the current session).
	TalliesGetter getters.Getter[[]Tally]

	// Keys is the list of Tally struct field names that can be charted,
	// e.g. []string{"Visits","Appointments","Leads","Presentations","Demos",
	// "Letters","FollowUps","Proposals","Policies","Premium"}.
	Keys []string
}

func (s StatLineChart) GetKey() string {
	return s.Key
}

func (s StatLineChart) GetRoles() []string {
	return s.Roles
}

func (s StatLineChart) GetChildren() []components.PageInterface {
	return nil
}

func (s *StatLineChart) SetChildren(children []components.PageInterface) {
	// no-op
}

func (s StatLineChart) Build(ctx context.Context) Node {
	// Resolve tallies via getter.
	if s.TalliesGetter == nil {
		slog.Error("StatLineChart: TalliesGetter is nil")
		return Div(Text("Error loading chart data"))
	}
	tallies, err := s.TalliesGetter(ctx)
	if err != nil {
		slog.Error("StatLineChart: failed to resolve tallies", "error", err)
		return Div(Text("Error loading chart data"))
	}

	// Prepare date labels (sorted unique dates).
	dateSet := map[string]struct{}{}
	for _, t := range tallies {
		dateSet[t.Date.Format("2006-01-02")] = struct{}{}
	}
	dates := make([]string, 0, len(dateSet))
	for d := range dateSet {
		dates = append(dates, d)
	}
	sort.Strings(dates)

	// Precompute per-date values for each key.
	// We build: map[key]map[date]value
	valuesByKey := map[string]map[string]int{}
	for _, key := range s.Keys {
		valuesByKey[key] = map[string]int{}
	}
	for _, t := range tallies {
		d := t.Date.Format("2006-01-02")
		for _, key := range s.Keys {
			switch key {
			case "Visits":
				valuesByKey[key][d] += t.Visits
			case "Appointments":
				valuesByKey[key][d] += t.Appointments
			case "Leads":
				valuesByKey[key][d] += t.Leads
			case "Presentations":
				valuesByKey[key][d] += t.Presentations
			case "Demos":
				valuesByKey[key][d] += t.Demos
			case "Letters":
				valuesByKey[key][d] += t.Letters
			case "FollowUps":
				valuesByKey[key][d] += t.FollowUps
			case "Proposals":
				valuesByKey[key][d] += t.Proposals
			case "Policies":
				valuesByKey[key][d] += t.Policies
			case "Premium":
				valuesByKey[key][d] += t.Premium
			default:
				// ignore unknown keys
			}
		}
	}

	// Build JSON-friendly slices for initial key (first key).
	initialKey := ""
	if len(s.Keys) > 0 {
		initialKey = s.Keys[0]
	}

	// We embed the date list and per-key values as data-* attributes to be
	// read by Alpine + ApexCharts on the client.
	return Div(
		Class("mt-4"),
		Div(Class("font-semibold mb-2"), Text("Tally Trend")),
		Div( // chart container with Alpine state
			// Alpine state: activeKey, dates, valuesByKey and chart init/update.
			Attr("x-data", fmt.Sprintf(`{
				activeKey: %q,
				dates: %s,
				valuesByKey: %s,
				chart: null,
				init() {
					const el = this.$refs.chart;
					const options = {
						chart: { type: 'line', height: 260, toolbar: { show: false } },
						stroke: { curve: 'smooth', width: 2 },
						xaxis: {
							categories: this.dates,
							labels: { show: false },
							axisTicks: { show: false },
							axisBorder: { show: false },
						},
						yaxis: {
							labels: { show: false },
						},
						series: [{ name: this.activeKey, data: this.valuesByKey[this.activeKey] || [] }],
						dataLabels: { enabled: false },
						tooltip: {
							custom: function({ series, seriesIndex, dataPointIndex }) {
								var val = series[seriesIndex][dataPointIndex];
								return '<div class="px-2 py-1 text-xs">' + (val != null ? val.toString() : '') + '</div>';
							},
						},
					};
					this.chart = new ApexCharts(el, options);
					this.chart.render();
					this.$watch('activeKey', (value) => {
						this.chart.updateSeries([{ name: value, data: this.valuesByKey[value] || [] }]);
					});
				}
			}`, initialKey, encodeStringSliceForJS(dates), encodeValuesByKeyForJS(s.Keys, dates, valuesByKey))),
			Div(
				Class("flex flex-wrap gap-2 mb-2"),
				// Key selector buttons
				func() Node {
					btns := Group{}
					for _, key := range s.Keys {
						label := key
						btns = append(btns,
							Button(
								Type("button"),
								Class("btn btn-xs"),
								Attr(":class", fmt.Sprintf(`activeKey === %q ? 'btn-primary text-primary-content' : 'btn-ghost'`, label)),
								Attr("@click", fmt.Sprintf("activeKey = %q", label)),
								Text(label),
							),
						)
					}
					return btns
				}(),
			),
			Div(Attr("x-ref", "chart"), Class("w-full h-64 bg-base-100 rounded-box shadow-inner")),
		),
	)
}

// Helpers to encode Go data into JS array literals for Alpine/ApexCharts.
func encodeStringSliceForJS(items []string) string {
	parts := make([]string, len(items))
	for i, v := range items {
		parts[i] = fmt.Sprintf("%q", v)
	}
	return "[" + strings.Join(parts, ",") + "]"
}

func encodeValuesByKeyForJS(keys, dates []string, valuesByKey map[string]map[string]int) string {
	// valuesByKeyJS will be: {key: [values aligned with dates], ...}
	pairs := make([]string, 0, len(keys))
	for _, key := range keys {
		vals := make([]string, len(dates))
		for i, d := range dates {
			vals[i] = fmt.Sprintf("%d", valuesByKey[key][d])
		}
		pairs = append(pairs, fmt.Sprintf("%q:[%s]", key, strings.Join(vals, ",")))
	}
	return "{" + strings.Join(pairs, ",") + "}"
}
