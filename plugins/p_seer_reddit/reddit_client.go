package p_seer_reddit

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
)

type redditObject[T redditObjectInterface] struct {
	Kind string `json:"kind"`
	Data T      `json:"data"`
}

type redditListing[T redditObjectInterface] struct {
	After    *string           `json:"after"`
	Dist     int               `json:"dist"`
	Modhash  string            `json:"modhash"`
	Children []redditObject[T] `json:"children"`
}

func (redditListing[T]) Kind() string {
	return "Listing"
}

type redditObjectInterface interface {
	Kind() string
}

// RedditPostData mirrors Reddit JSON "t3" children.
type RedditPostData struct {
	ID             string  `json:"id"`
	Name           string  `json:"name"`
	Subreddit      string  `json:"subreddit"`
	SubredditID    string  `json:"subreddit_id"`
	SubredditType  string  `json:"subreddit_type"`
	Author         string  `json:"author"`
	AuthorFullname string  `json:"author_fullname"`
	Title          string  `json:"title"`
	Selftext       string  `json:"selftext"`
	SelftextHTML   string  `json:"selftext_html"`
	URL            string  `json:"url"`
	Permalink      string  `json:"permalink"`
	CreatedUTC     float64 `json:"created_utc"`
	Edited         any     `json:"edited"`
	Score          int     `json:"score"`
	Ups            int     `json:"ups"`
	Downs          int     `json:"downs"`
	NumComments    int     `json:"num_comments"`
	IsSelf         bool    `json:"is_self"`
	IsVideo        bool    `json:"is_video"`
	Domain         string  `json:"domain"`
	Thumbnail      string  `json:"thumbnail"`
	Over18         bool    `json:"over_18"`
	Spoiler        bool    `json:"spoiler"`
	Stickied       bool    `json:"stickied"`
	Locked         bool    `json:"locked"`
	Archived       bool    `json:"archived"`
	RemovedBy      *string `json:"removed_by_category"`
}

func (RedditPostData) Kind() string {
	return "t3"
}

const redditListingMaxPages = 25

func fetchSubredditPosts(ctx context.Context, subreddit string, after *string) (*redditObject[redditListing[RedditPostData]], error) {
	base := fmt.Sprintf("https://www.reddit.com/r/%s/.json", url.PathEscape(subreddit))
	v := url.Values{}
	if after != nil && strings.TrimSpace(*after) != "" {
		v.Set("after", *after)
	}
	u := base
	if len(v) > 0 {
		u = base + "?" + v.Encode()
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		slog.Error("p_seer_reddit: new request", "error", err, "subreddit", subreddit)
		return nil, err
	}
	req.Header.Set("User-Agent", "lago:p_seer_reddit:1.0 (by /u/local)")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		slog.Error("p_seer_reddit: fetch listing", "error", err, "subreddit", subreddit)
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		err := fmt.Errorf("reddit: %s", resp.Status)
		slog.Error("p_seer_reddit: fetch listing", "error", err, "subreddit", subreddit)
		return nil, err
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		slog.Error("p_seer_reddit: read listing body", "error", err, "subreddit", subreddit)
		return nil, err
	}
	var out redditObject[redditListing[RedditPostData]]
	if err := json.Unmarshal(body, &out); err != nil {
		slog.Error("p_seer_reddit: decode listing", "error", err, "subreddit", subreddit)
		return nil, err
	}
	return new(out), nil
}

func fetchSubredditPostsSearch(ctx context.Context, subreddit, query string, after *string) (*redditObject[redditListing[RedditPostData]], error) {
	base := fmt.Sprintf("https://www.reddit.com/r/%s/search.json", url.PathEscape(subreddit))
	q := url.Values{}
	q.Set("q", query)
	q.Set("restrict_sr", "1")
	if after != nil && strings.TrimSpace(*after) != "" {
		q.Set("after", *after)
	}
	u := base + "?" + q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		slog.Error("p_seer_reddit: new request (search)", "error", err, "subreddit", subreddit)
		return nil, err
	}
	req.Header.Set("User-Agent", "lago:p_seer_reddit:1.0 (by /u/local)")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		slog.Error("p_seer_reddit: fetch search", "error", err, "subreddit", subreddit)
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		err := fmt.Errorf("reddit: %s", resp.Status)
		slog.Error("p_seer_reddit: fetch search", "error", err, "subreddit", subreddit)
		return nil, err
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		slog.Error("p_seer_reddit: read search body", "error", err, "subreddit", subreddit)
		return nil, err
	}
	var out redditObject[redditListing[RedditPostData]]
	if err := json.Unmarshal(body, &out); err != nil {
		slog.Error("p_seer_reddit: decode search", "error", err, "subreddit", subreddit)
		return nil, err
	}
	return new(out), nil
}
