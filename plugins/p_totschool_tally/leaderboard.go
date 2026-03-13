package p_totschool_tally

import (
	"context"
	"fmt"
	"strings"

	"github.com/lariv-in/components"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

type TallyLeaderboardComponent struct {
	components.Page
}

func (l TallyLeaderboardComponent) Build(ctx context.Context) Node {
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

func (l TallyLeaderboardComponent) GetChildren() []components.PageInterface {
	return nil
}
