package p_seer_reddit

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_seer_intel"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

// redditPostBulkAddIntelFormURL is the POST target for bulk intel ingest; preserves the list URL query string (e.g. page).
func redditPostBulkAddIntelFormURL(routeName string, pathParams map[string]getters.Getter[any]) getters.Getter[string] {
	return func(ctx context.Context) (string, error) {
		base, err := lago.RoutePath(routeName, pathParams)(ctx)
		if err != nil {
			return "", err
		}
		reqVal := ctx.Value("$request")
		r, ok := reqVal.(*http.Request)
		if !ok || r == nil || r.URL == nil || r.URL.RawQuery == "" {
			return base, nil
		}
		return base + "?" + r.URL.RawQuery, nil
	}
}

// redditPostIntelMissingGetter is true when no [p_seer_intel.Intel] row exists yet for this post (kind + KindID).
func redditPostIntelMissingGetter() getters.Getter[any] {
	return func(ctx context.Context) (any, error) {
		post, err := getters.Key[RedditPost]("redditPost")(ctx)
		if err != nil {
			return false, err
		}
		if post.ID == 0 {
			return false, nil
		}
		db, err := getters.DBFromContext(ctx)
		if err != nil {
			return false, err
		}
		exists, err := p_seer_intel.IntelExistsForSource(ctx, db, (RedditPost{}).Kind(), post.ID)
		if err != nil {
			return false, err
		}
		return !exists, nil
	}
}

// redditPostIntelPresentGetter is true when an [p_seer_intel.Intel] row exists for this post.
func redditPostIntelPresentGetter() getters.Getter[any] {
	return func(ctx context.Context) (any, error) {
		post, err := getters.Key[RedditPost]("redditPost")(ctx)
		if err != nil {
			return false, err
		}
		if post.ID == 0 {
			return false, nil
		}
		db, err := getters.DBFromContext(ctx)
		if err != nil {
			return false, err
		}
		ok, err := p_seer_intel.IntelExistsForSource(ctx, db, (RedditPost{}).Kind(), post.ID)
		return ok, err
	}
}

// redditPostIntelDetailHrefGetter returns the app path to [seer_intel.DetailRoute] for intel linked to this post.
func redditPostIntelDetailHrefGetter() getters.Getter[string] {
	return func(ctx context.Context) (string, error) {
		post, err := getters.Key[RedditPost]("redditPost")(ctx)
		if err != nil {
			return "", err
		}
		return p_seer_intel.IntelDetailPathForSource(ctx, (RedditPost{}).Kind(), post.ID)
	}
}

// redditPostToolbarBusyGetter disables Reddit post toolbar POSTs while intel ingest or async source fetch runs.
func redditPostToolbarBusyGetter() getters.Getter[Node] {
	return func(context.Context) (Node, error) {
		if redditIntelIngestActive.Load() || redditFetchPostsActive.Load() {
			return Group{Disabled(), Class("btn-disabled")}, nil
		}
		return nil, nil
	}
}

func redditPostListViewPollURL(ctx context.Context, bySource bool, sourceID uint) (string, error) {
	routeName := "seer_reddit.RedditPostListRoute"
	var pathArgs map[string]getters.Getter[any]
	if bySource {
		routeName = "seer_reddit.RedditPostListBySourceRoute"
		pathArgs = map[string]getters.Getter[any]{
			"source_id": getters.Any(getters.Static(strconv.FormatUint(uint64(sourceID), 10))),
		}
	}
	base, err := lago.RoutePath(routeName, pathArgs)(ctx)
	if err != nil {
		return "", err
	}
	reqVal := ctx.Value("$request")
	r, ok := reqVal.(*http.Request)
	if !ok || r == nil || r.URL == nil || r.URL.RawQuery == "" {
		return base, nil
	}
	return base + "?" + r.URL.RawQuery, nil
}

func redditPostDetailPollURL(ctx context.Context, postID uint) (string, error) {
	base, err := lago.RoutePath("seer_reddit.RedditPostDetailRoute", map[string]getters.Getter[any]{
		"id": getters.Any(getters.Static(strconv.FormatUint(uint64(postID), 10))),
	})(ctx)
	if err != nil {
		return "", err
	}
	reqVal := ctx.Value("$request")
	r, ok := reqVal.(*http.Request)
	if !ok || r == nil || r.URL == nil || r.URL.RawQuery == "" {
		return base, nil
	}
	return base + "?" + r.URL.RawQuery, nil
}

// redditPostTableHasSourceContext is true when `redditSource.ID` is available (by-source post list shell).
func redditPostTableHasSourceContext() getters.Getter[bool] {
	return func(ctx context.Context) (bool, error) {
		_, err := getters.Key[uint]("redditSource.ID")(ctx)
		return err == nil, nil
	}
}

// newRedditPostDataTable builds the shared Reddit post [components.DataTable] for global and by-source lists;
// toolbar actions that differ by shell use [components.ShowIf].
func newRedditPostDataTable() *components.DataTable[RedditPost] {
	sourceScoped := redditPostTableHasSourceContext()
	sourceID := getters.Any(getters.Key[uint]("redditSource.ID"))
	return &components.DataTable[RedditPost]{
		Page:    components.Page{Key: "seer_reddit.RedditPostTableBody"},
		UID:     "seer-reddit-posts-table",
		Classes: "w-full",
		Data:    getters.Key[components.ObjectList[RedditPost]]("redditPosts"),
		Actions: []components.PageInterface{
			&components.ShowIf{
				Getter: getters.Any(sourceScoped),
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
				Getter: getters.BoolNot(sourceScoped),
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

func redditPostListTableShellGetter(bySource bool) getters.Getter[components.PageInterface] {
	return func(ctx context.Context) (components.PageInterface, error) {
		var sourceID uint
		if bySource {
			sid, err := getters.Key[uint]("redditSource.ID")(ctx)
			if err != nil {
				return nil, err
			}
			sourceID = sid
		}
		tbl := newRedditPostDataTable()
		busy := redditIntelIngestActive.Load()
		if bySource {
			busy = busy || redditFetchPostsActive.Load()
		}
		if !busy {
			return tbl, nil
		}
		u, err := redditPostListViewPollURL(ctx, bySource, sourceID)
		if err != nil {
			return nil, err
		}
		return &components.HTMXPolling{
			Page:     components.Page{Key: "seer_reddit.RedditPostTablePolling"},
			URL:      getters.Static(u),
			Children: []components.PageInterface{tbl},
		}, nil
	}
}

func redditPostDetailShellGetter(inner components.PageInterface) getters.Getter[components.PageInterface] {
	return func(ctx context.Context) (components.PageInterface, error) {
		post, err := getters.Key[RedditPost]("redditPost")(ctx)
		if err != nil {
			return nil, err
		}
		if !redditIntelIngestActive.Load() {
			return inner, nil
		}
		u, err := redditPostDetailPollURL(ctx, post.ID)
		if err != nil {
			return nil, err
		}
		return &components.HTMXPolling{
			Page:     components.Page{Key: "seer_reddit.RedditPostDetailPolling"},
			URL:      getters.Static(u),
			Children: []components.PageInterface{inner},
		}, nil
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
						Getter: redditPostIntelMissingGetter(),
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
			&components.ShowIf{
				Page:   components.Page{Key: "seer_reddit.RedditPostDetailIntelLinkWrap"},
				Getter: getters.Any(redditPostIntelPresentGetter()),
				Children: []components.PageInterface{
					&components.LabelInline{
						Title: "Intel",
						Children: []components.PageInterface{
							&components.FieldLink{
								Page:    components.Page{Key: "seer_reddit.RedditPostDetailIntelLink"},
								Href:    redditPostIntelDetailHrefGetter(),
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
