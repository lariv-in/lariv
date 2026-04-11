package p_lacerate

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"html"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type RedditObject[T RedditObjectInterface] struct {
	Kind string `json:"kind"`
	Data T      `json:"data"`
}

type RedditListing[T RedditObjectInterface] struct {
	After     *string           `json:"after"`
	Dist      int               `json:"dist"`
	Modhash   string            `json:"modhash"`
	GeoFilter *string           `json:"geo_filter"`
	Children  []RedditObject[T] `json:"children"`
}

func (RedditListing[T]) Kind() string {
	return "Listing"
}

type RedditObjectInterface interface {
	Kind() string
}

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

// IntelDedupHash returns a stable dedupe key from the Reddit listing JSON `id` (short post id).
func (p RedditPostData) IntelDedupHash() string {
	id := strings.TrimSpace(p.ID)
	if id == "" {
		return ""
	}
	sum := sha256.Sum256([]byte(id))
	return hex.EncodeToString(sum[:])
}

// PreviewImageURL returns the post thumbnail when it is an absolute http(s) URL.
func (p RedditPostData) PreviewImageURL() string {
	t := strings.TrimSpace(html.UnescapeString(p.Thumbnail))
	u, err := url.Parse(t)
	if err != nil {
		return ""
	}
	if (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		return ""
	}
	return t
}

// Markdown builds markdown body for ingest into [Intel.Content].
func (p RedditPostData) Markdown(ctx context.Context) string {
	var b strings.Builder
	title := strings.TrimSpace(p.Title)
	if title != "" {
		b.WriteString("# ")
		b.WriteString(title)
		b.WriteString("\n\n")
	}
	if txt := strings.TrimSpace(p.Selftext); txt != "" {
		b.WriteString(txt)
		b.WriteString("\n\n")
	}
	if !p.IsSelf {
		if u := strings.TrimSpace(p.URL); u != "" {
			if body := fetchPostURLAsMarkdown(ctx, u); body != "" {
				b.WriteString("## Linked article\n\n")
				b.WriteString(body)
				b.WriteString("\n\n")
			}
		}
	}
	b.WriteString("---\n\n")
	fmt.Fprintf(&b, "- **Author:** u/%s\n", p.Author)
	fmt.Fprintf(&b, "- **Subreddit:** r/%s\n", p.Subreddit)
	if p.CreatedUTC > 0 {
		fmt.Fprintf(&b, "- **Posted:** %s\n", time.Unix(int64(p.CreatedUTC), 0).UTC().Format(time.RFC3339))
	}
	fmt.Fprintf(&b, "- **Score:** %d (up %d / down %d)\n", p.Score, p.Ups, p.Downs)
	fmt.Fprintf(&b, "- **Comments:** %d\n", p.NumComments)
	if p.Permalink != "" {
		fmt.Fprintf(&b, "- **Permalink:** https://www.reddit.com%s\n", p.Permalink)
	}
	if !p.IsSelf {
		if u := strings.TrimSpace(p.URL); u != "" {
			fmt.Fprintf(&b, "- **Link:** %s\n", u)
		}
	}
	return strings.TrimSpace(b.String())
}

func FetchSubredditPosts(subreddit string, after *string) (*RedditObject[RedditListing[RedditPostData]], error) {
	base := fmt.Sprintf("https://www.reddit.com/r/%s/.json", url.PathEscape(subreddit))
	v := url.Values{}
	if after != nil && strings.TrimSpace(*after) != "" {
		v.Set("after", *after)
	}
	u := base
	if len(v) > 0 {
		u = base + "?" + v.Encode()
	}
	req, err := http.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		slog.Error("lacerate: reddit new request", "error", err, "subreddit", subreddit)
		return nil, err
	}
	req.Header.Set("User-Agent", "lago:p_lacerate:1.0 (by /u/local)")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		slog.Error("lacerate: reddit fetch listing", "error", err, "subreddit", subreddit)
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		err := fmt.Errorf("reddit: %s", resp.Status)
		slog.Error("lacerate: reddit fetch listing", "error", err, "subreddit", subreddit)
		return nil, err
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		slog.Error("lacerate: reddit read listing body", "error", err, "subreddit", subreddit)
		return nil, err
	}
	var out RedditObject[RedditListing[RedditPostData]]
	if err := json.Unmarshal(body, &out); err != nil {
		slog.Error("lacerate: reddit decode listing", "error", err, "subreddit", subreddit)
		return nil, err
	}
	return &out, nil
}

func FetchSubredditPostsSearch(subreddit, query string, after *string) (*RedditObject[RedditListing[RedditPostData]], error) {
	base := fmt.Sprintf("https://www.reddit.com/r/%s/search.json", url.PathEscape(subreddit))
	q := url.Values{}
	q.Set("q", query)
	q.Set("restrict_sr", "1") // keep results in this subreddit
	if after != nil && strings.TrimSpace(*after) != "" {
		q.Set("after", *after)
	}
	u := base + "?" + q.Encode()

	req, err := http.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		slog.Error("lacerate: reddit new request (search)", "error", err, "subreddit", subreddit)
		return nil, err
	}
	req.Header.Set("User-Agent", "lago:p_lacerate:1.0 (by /u/local)")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		slog.Error("lacerate: reddit fetch search", "error", err, "subreddit", subreddit)
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		err := fmt.Errorf("reddit: %s", resp.Status)
		slog.Error("lacerate: reddit fetch search", "error", err, "subreddit", subreddit)
		return nil, err
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		slog.Error("lacerate: reddit read search body", "error", err, "subreddit", subreddit)
		return nil, err
	}
	var out RedditObject[RedditListing[RedditPostData]]
	if err := json.Unmarshal(body, &out); err != nil {
		slog.Error("lacerate: reddit decode search", "error", err, "subreddit", subreddit)
		return nil, err
	}
	return &out, nil
}
