package p_seer_reddit

import (
	"context"
	"time"

	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
)

// newRedditPostDataTable builds the shared Reddit post [components.DataTable] for global and by-source lists;
// toolbar actions that differ by shell use [components.ShowIf] with [redditPostListBySource] from the view layer.
func newRedditPostDataTable() *components.DataTable[RedditPost] {
	bySource := getters.Key[bool]("redditPostListBySource")
	sourceID := getters.Any(getters.Key[uint]("redditSource.ID"))
	return &components.DataTable[RedditPost]{
		Page:    components.Page{Key: "seer_reddit.RedditPostTableBody"},
		UID:     "seer-reddit-posts-table",
		Classes: "w-full",
		Data:    getters.Key[components.ObjectList[RedditPost]]("redditPosts"),
		Actions: []components.PageInterface{
			&components.ShowIf{
				Getter: getters.Any(bySource),
				Children: []components.PageInterface{
					&components.ContainerRow{
						Page:    components.Page{Key: "seer_reddit.RedditPostTableSourceActionsRow"},
						Classes: "items-center shrink-0",
						Children: []components.PageInterface{
							&components.ButtonPost{
								Page:    components.Page{Key: "seer_reddit.RedditPostTableFetchFromReddit"},
								Label:   "Load",
								URL:     lago.RoutePath("seer_reddit.RedditSourceFetchPostsRoute", map[string]getters.Getter[any]{"source_id": sourceID}),
								Icon:    "arrow-path",
								Classes: "btn-outline btn-sm w-24",
								Attr:    redditPostToolbarBusyGetter(),
							},
							&components.ButtonPost{
								Page:    components.Page{Key: "seer_reddit.RedditPostTableBySourceBulkAddIntel"},
								Label:   "Add to Intel",
								URL:     redditPostBulkAddIntelFormURL("seer_reddit.RedditPostListBySourceBulkAddIntelRoute", map[string]getters.Getter[any]{"source_id": sourceID}),
								Icon:    "document-plus",
								Classes: "btn-outline btn-primary btn-sm shrink-0 w-32",
								Attr:    redditPostToolbarBusyGetter(),
							},
						},
					},
				},
			},
			&components.ShowIf{
				Getter: getters.BoolNot(bySource),
				Children: []components.PageInterface{
					&components.ButtonPost{
						Page:    components.Page{Key: "seer_reddit.RedditPostTableBulkAddIntel"},
						Label:   "Add to Intel",
						URL:     redditPostBulkAddIntelFormURL("seer_reddit.RedditPostListBulkAddIntelRoute", nil),
						Icon:    "document-plus",
						Classes: "btn-outline btn-primary btn-sm shrink-0 w-32",
						Attr:    redditPostToolbarBusyGetter(),
					},
				},
			},
		},
		RowAttr: getters.RowAttrNavigate(
			lago.RoutePath("seer_reddit.RedditPostDetailRoute", map[string]getters.Getter[any]{
				"id": getters.Any(getters.Key[uint]("$row.ID")),
			}),
		),
		Columns: redditPostListTableColumns(),
	}
}

// redditPostDetailContentColumn is the main column for [seer_reddit.RedditPostDetail].
func redditPostDetailContentColumn() components.PageInterface {
	return components.ContainerColumn{
		Page: components.Page{Key: "seer_reddit.RedditPostDetailContent"},
		Children: []components.PageInterface{
			&components.ContainerRow{
				Page:    components.Page{Key: "seer_reddit.RedditPostDetailHeader"},
				Classes: "flex flex-wrap justify-between items-start gap-2 w-full",
				Children: []components.PageInterface{
					&components.FieldTitle{Getter: getters.Key[string]("$in.Title")},
					&components.ShowIf{
						Page:   components.Page{Key: "seer_reddit.RedditPostDetailAddIntelWrap"},
						Getter: getters.Any(getters.Key[bool]("redditPostIntelAddVisible")),
						Children: []components.PageInterface{
							&components.ButtonPost{
								Page:    components.Page{Key: "seer_reddit.RedditPostDetailAddIntelBtn"},
								Label:   "Add to Intel",
								URL:     lago.RoutePath("seer_reddit.RedditPostAddIntelRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("$in.ID"))}),
								Icon:    "document-plus",
								Classes: "btn-outline btn-primary btn-sm shrink-0",
								Attr:    redditPostToolbarBusyGetter(),
							},
						},
					},
					&components.ButtonModalForm{
						Page:        components.Page{Key: "seer_reddit.RedditPostDetailDeleteBtn"},
						Label:       "Delete",
						Icon:        "trash",
						Name:        getters.Static("seer_reddit.RedditPostDeleteForm"),
						Url:         lago.RoutePath("seer_reddit.RedditPostDeleteRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("redditPost.ID"))}),
						FormPostURL: lago.RoutePath("seer_reddit.RedditPostDeleteRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("redditPost.ID"))}),
						ModalUID:    "seer-reddit-post-delete-modal",
						Classes:     "btn-outline btn-error btn-sm shrink-0",
					},
				},
			},
			&components.LabelInline{
				Title: "Post ID",
				Children: []components.PageInterface{
					&components.FieldText{Getter: getters.Key[string]("$in.PostID")},
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
						Getter: getters.Map(getters.Key[float64]("$in.CreatedUTCUnix"), func(_ context.Context, u float64) (string, error) {
							return time.Unix(int64(u), 0).UTC().Format(time.RFC3339), nil
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
			&components.ShowIf{
				Page:   components.Page{Key: "seer_reddit.RedditPostDetailIntelLinkWrap"},
				Getter: getters.Any(getters.Key[bool]("redditPostIntelLinkVisible")),
				Children: []components.PageInterface{
					&components.LabelInline{
						Title: "Intel",
						Children: []components.PageInterface{
							&components.FieldLink{
								Page:    components.Page{Key: "seer_reddit.RedditPostDetailIntelLink"},
								Href:    getters.Key[string]("redditPostIntelDetailHref"),
								Label:   getters.Static("View intel"),
								Classes: "link link-primary",
							},
						},
					},
				},
			},
		},
	}
}
