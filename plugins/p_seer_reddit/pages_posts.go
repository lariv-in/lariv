package p_seer_reddit

import (
	"context"
	"strconv"
	"time"

	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
)

func registerRedditPostPages() {
	lago.RegistryPage.Register("seer_reddit.RedditPostTable", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "seer_reddit.RedditMenu"},
		},
		Children: []components.PageInterface{
			&components.DataTable[RedditPost]{
				Page:    components.Page{Key: "seer_reddit.RedditPostTableBody"},
				UID:     "seer-reddit-posts-table",
				Classes: "w-full",
				Data:    getters.Key[components.ObjectList[RedditPost]]("redditPosts"),
				RowAttr: getters.RowAttrNavigate(
					lago.RoutePath("seer_reddit.RedditPostDetailRoute", map[string]getters.Getter[any]{
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
						Label: "Post ID",
						Children: []components.PageInterface{
							&components.FieldText{Getter: getters.Key[string]("$row.PostID")},
						},
					},
					{
						Label: "Title",
						Children: []components.PageInterface{
							&components.FieldText{Getter: getters.Key[string]("$row.Title")},
						},
					},
					{
						Label: "r/",
						Children: []components.PageInterface{
							&components.FieldText{Getter: getters.Key[string]("$row.Subreddit")},
						},
					},
					{
						Label: "Intel ID",
						Children: []components.PageInterface{
							&components.FieldText{
								Getter: getters.Map(getters.Key[RedditPost]("$row"), func(_ context.Context, row RedditPost) (string, error) {
									if row.IntelID == nil || *row.IntelID == 0 {
										return "—", nil
									}
									return strconv.FormatUint(uint64(*row.IntelID), 10), nil
								}),
							},
						},
					},
				},
			},
		},
	})

	lago.RegistryPage.Register("seer_reddit.RedditPostDetail", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "seer_reddit.RedditPostDetailMenu"},
		},
		Children: []components.PageInterface{
			&components.Detail[RedditPost]{
				Getter: getters.Key[RedditPost]("redditPost"),
				Children: []components.PageInterface{
					components.ContainerColumn{
						Page: components.Page{Key: "seer_reddit.RedditPostDetailContent"},
						Children: []components.PageInterface{
							&components.FieldTitle{Getter: getters.Key[string]("$in.Title")},
							&components.LabelInline{
								Title: "Post ID",
								Children: []components.PageInterface{
									&components.FieldText{Getter: getters.Key[string]("$in.PostID")},
								},
							},
							&components.LabelInline{
								Title: "Intel ID",
								Children: []components.PageInterface{
									&components.FieldText{
										Getter: getters.Map(getters.Key[RedditPost]("redditPost"), func(_ context.Context, p RedditPost) (string, error) {
											if p.IntelID == nil || *p.IntelID == 0 {
												return "—", nil
											}
											return strconv.FormatUint(uint64(*p.IntelID), 10), nil
										}),
									},
								},
							},
							&components.LabelInline{
								Title: "Subreddit",
								Children: []components.PageInterface{
									&components.FieldText{Getter: getters.Format("r/%s", getters.Any(getters.Key[string]("$in.Subreddit")))},
								},
							},
							&components.LabelInline{
								Title: "Author",
								Children: []components.PageInterface{
									&components.FieldText{Getter: getters.Format("u/%s", getters.Any(getters.Key[string]("$in.Author")))},
								},
							},
							&components.LabelInline{
								Title: "Created (UTC)",
								Children: []components.PageInterface{
									&components.FieldText{
										Getter: getters.Map(getters.Key[RedditPost]("redditPost"), func(_ context.Context, p RedditPost) (string, error) {
											return time.Unix(int64(p.CreatedUTCUnix), 0).UTC().Format(time.RFC3339), nil
										}),
									},
								},
							},
							&components.LabelInline{
								Title: "Selftext",
								Children: []components.PageInterface{
									&components.FieldText{
										Getter:  getters.Key[string]("$in.Selftext"),
										Classes: "whitespace-pre-wrap",
									},
								},
							},
							&components.LabelInline{
								Title: "IntelKind Content()",
								Children: []components.PageInterface{
									&components.FieldText{
										Getter: getters.Map(getters.Key[RedditPost]("redditPost"), func(_ context.Context, p RedditPost) (string, error) {
											return p.Content(), nil
										}),
										Classes: "whitespace-pre-wrap font-mono text-sm",
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
