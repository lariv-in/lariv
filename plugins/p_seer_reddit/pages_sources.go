package p_seer_reddit

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	"gorm.io/datatypes"
)

// formatSubredditsJSON turns stored JSON array of names into comma-separated text.
// If maxNames > 0, only the first maxNames entries are shown and ", …" is appended when the list is longer.
func formatSubredditsJSON(raw []byte, maxNames int) string {
	if len(raw) == 0 {
		return "—"
	}
	var subs []string
	if err := json.Unmarshal(raw, &subs); err != nil {
		return string(raw)
	}
	if len(subs) == 0 {
		return "—"
	}
	out := subs
	truncated := false
	if maxNames > 0 && len(out) > maxNames {
		out = out[:maxNames]
		truncated = true
	}
	var s strings.Builder
	for i, x := range out {
		if i > 0 {
			s.WriteString(", ")
		}
		s.WriteString(x)
	}
	if truncated {
		s.WriteString(", …")
	}
	return s.String()
}

func subredditsPreview(ctx context.Context) (string, error) {
	rs, err := getters.Key[RedditSource]("redditSource")(ctx)
	if err != nil {
		return "", err
	}
	return formatSubredditsJSON([]byte(rs.Subreddits), 5), nil
}

func subredditsBytesFromRowValue(v any) []byte {
	if v == nil {
		return nil
	}
	switch x := v.(type) {
	case []byte:
		return x
	case string:
		return []byte(x)
	case datatypes.JSON:
		return []byte(x)
	default:
		return nil
	}
}

func subredditsFromTableRow(ctx context.Context) (string, error) {
	rowAny := ctx.Value("$row")
	m, ok := rowAny.(map[string]any)
	if !ok {
		return "—", nil
	}
	return formatSubredditsJSON(subredditsBytesFromRowValue(m["Subreddits"]), 0), nil
}

func redditSourceLoadWebsitesYesNoFromRow(ctx context.Context) (string, error) {
	rowAny := ctx.Value("$row")
	m, ok := rowAny.(map[string]any)
	if !ok {
		return "—", nil
	}
	v, ok := m["LoadWebsites"]
	if !ok {
		return "No", nil
	}
	switch b := v.(type) {
	case bool:
		if b {
			return "Yes", nil
		}
		return "No", nil
	default:
		return "—", nil
	}
}

func redditSourceLoadWebsitesYesNoFromDetail(ctx context.Context) (string, error) {
	b, err := getters.Key[bool]("$in.LoadWebsites")(ctx)
	if err != nil {
		return "", err
	}
	if b {
		return "Yes", nil
	}
	return "No", nil
}

func redditSourceDetailWorkerLabel(ctx context.Context) (string, error) {
	rs, err := getters.Key[RedditSource]("redditSource")(ctx)
	if err != nil {
		return "", err
	}
	if rs.RedditRunnerID == nil || *rs.RedditRunnerID == 0 {
		return "—", nil
	}
	if rs.RedditRunner != nil {
		return rs.RedditRunner.Name, nil
	}
	db, err := getters.DBFromContext(ctx)
	if err != nil {
		return "", err
	}
	var rr RedditRunner
	if err := db.WithContext(ctx).Where("id = ?", *rs.RedditRunnerID).Take(&rr).Error; err != nil {
		return fmt.Sprintf("id %d", *rs.RedditRunnerID), nil
	}
	return rr.Name, nil
}

func registerRedditSourcePages() {
	lago.RegistryPage.Register("seer_reddit.RedditSourceTable", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "seer_reddit.RedditMenu"},
		},
		Children: []components.PageInterface{
			&components.DataTable[RedditSource]{
				Page:    components.Page{Key: "seer_reddit.RedditSourceTableBody"},
				UID:     "seer-reddit-sources-table",
				Classes: "w-full",
				Data:    getters.Key[components.ObjectList[RedditSource]]("redditSources"),
				Actions: []components.PageInterface{
					&components.TableButtonCreate{Link: lago.RoutePath("seer_reddit.RedditSourceCreateRoute", nil)},
				},
				RowAttr: getters.RowAttrNavigate(
					lago.RoutePath("seer_reddit.RedditSourceDetailRoute", map[string]getters.Getter[any]{
						"id": getters.Any(getters.Key[uint]("$row.ID")),
					}),
				),
				Columns: []components.TableColumn{
					{
						Label: "Subreddits",
						Children: []components.PageInterface{
							&components.FieldText{Getter: subredditsFromTableRow},
						},
					},
					{
						Label: "Search",
						Children: []components.PageInterface{
							&components.FieldText{Getter: getters.Key[string]("$row.SearchQuery")},
						},
					},
					{
						Label: "Max fresh",
						Children: []components.PageInterface{
							&components.FieldText{Getter: getters.Format("%d", getters.Any(getters.Key[uint]("$row.MaxFreshPosts")))},
						},
					},
					{
						Label: "Websites",
						Children: []components.PageInterface{
							&components.FieldText{Getter: redditSourceLoadWebsitesYesNoFromRow},
						},
					},
				},
			},
		},
	})

	lago.RegistryPage.Register("seer_reddit.RedditSourceDetail", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "seer_reddit.RedditSourceDetailMenu"},
		},
		Children: []components.PageInterface{
			&components.Detail[RedditSource]{
				Getter: getters.Key[RedditSource]("redditSource"),
				Children: []components.PageInterface{
					components.ContainerColumn{
						Page: components.Page{Key: "seer_reddit.RedditSourceDetailContent"},
						Children: []components.PageInterface{
							&components.FieldTitle{
								Getter: getters.Format("Source #%d", getters.Any(getters.Key[uint]("$in.ID"))),
							},
							&components.LabelInline{
								Title: "Worker",
								Children: []components.PageInterface{
									&components.FieldText{
										Getter: redditSourceDetailWorkerLabel,
									},
								},
							},
							&components.LabelInline{
								Title: "Subreddits",
								Children: []components.PageInterface{
									&components.FieldText{Getter: subredditsPreview},
								},
							},
							&components.LabelInline{
								Title: "Search query",
								Children: []components.PageInterface{
									&components.FieldText{Getter: getters.Key[string]("$in.SearchQuery")},
								},
							},
							&components.LabelInline{
								Title: "Max fresh posts",
								Children: []components.PageInterface{
									&components.FieldText{Getter: getters.Format("%d", getters.Any(getters.Key[uint]("$in.MaxFreshPosts")))},
								},
							},
							&components.LabelInline{
								Title: "Load websites",
								Children: []components.PageInterface{
									&components.FieldText{Getter: redditSourceLoadWebsitesYesNoFromDetail},
								},
							},
						},
					},
				},
			},
		},
	})
}
