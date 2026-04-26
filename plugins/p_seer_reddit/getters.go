package p_seer_reddit

import (
	"context"
	"net/http"
	"strconv"

	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
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
