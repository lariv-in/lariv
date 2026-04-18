package p_seer_reddit

import (
	"context"
	"encoding/json"
	"strconv"

	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
)

func subredditsPreview(ctx context.Context) (string, error) {
	rs, err := getters.Key[RedditSource]("redditSource")(ctx)
	if err != nil {
		return "", err
	}
	raw := []byte(rs.Subreddits)
	if len(raw) == 0 {
		return "—", nil
	}
	var subs []string
	if err := json.Unmarshal(raw, &subs); err != nil {
		return string(raw), nil
	}
	if len(subs) == 0 {
		return "—", nil
	}
	out := subs
	if len(out) > 5 {
		out = out[:5]
	}
	s := ""
	for i, x := range out {
		if i > 0 {
			s += ", "
		}
		s += x
	}
	if len(subs) > 5 {
		s += ", …"
	}
	return s, nil
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
				RowAttr: getters.RowAttrNavigate(
					lago.RoutePath("seer_reddit.RedditSourceDetailRoute", map[string]getters.Getter[any]{
						"id": getters.Any(getters.Key[uint]("$row.ID")),
					}),
				),
				Columns: []components.TableColumn{
					{
						Label: "ID",
						Children: []components.PageInterface{
							&components.FieldText{Getter: getters.Format("%d", getters.Any(getters.Key[uint]("$row.ID")))},
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
						Label: "Runner ID",
						Children: []components.PageInterface{
							&components.FieldText{
								Getter: getters.Map(getters.Key[*uint]("$row.RunnerID"), func(_ context.Context, id *uint) (string, error) {
									if id == nil || *id == 0 {
										return "—", nil
									}
									return strconv.FormatUint(uint64(*id), 10), nil
								}),
							},
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
								Title: "Runner ID",
								Children: []components.PageInterface{
									&components.FieldText{
										Getter: getters.Map(getters.Key[RedditSource]("redditSource"), func(_ context.Context, rs RedditSource) (string, error) {
											if rs.RunnerID == nil || *rs.RunnerID == 0 {
												return "—", nil
											}
											return strconv.FormatUint(uint64(*rs.RunnerID), 10), nil
										}),
									},
								},
							},
						},
					},
				},
			},
		},
	})
}
