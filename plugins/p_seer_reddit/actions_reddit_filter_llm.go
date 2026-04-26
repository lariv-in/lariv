package p_seer_reddit

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/lariv-in/lago/plugins/p_google_genai"
	"google.golang.org/genai"
	"gorm.io/gorm"
)

// redditFilterLLMResult is the JSON object from the filter model (field "pass").
type redditFilterLLMResult struct {
	Pass bool `json:"pass"`
}

const redditFilterUserPromptMaxRunes = 12000

func trimRunes(s string, max int) string {
	if max <= 0 || s == "" {
		return s
	}
	if utf8.RuneCountInString(s) <= max {
		return s
	}
	r := []rune(s)
	if len(r) > max {
		return string(r[:max])
	}
	return s
}

// redditSourceUsesFilterLlm is true when [RedditSource] has non-empty [RedditSource.Filter] and
// new posts must be classified by the model before a full row is stored.
func redditSourceUsesFilterLlm(r *RedditSource) bool {
	if r == nil {
		return false
	}
	return strings.TrimSpace(r.Filter) != ""
}

// redditLLMPassesFilter asks Gemini (structured JSON) whether the post should be ingested
// for this source, given the filter and whitelist/blacklist mode.
func redditLLMPassesFilter(ctx context.Context, src *RedditSource, post RedditPostData) (bool, string, error) {
	if !redditSourceUsesFilterLlm(src) {
		return true, "", nil
	}
	if ctx.Err() != nil {
		return false, "", ctx.Err()
	}
	sys := `You are a strict classifier for Reddit post ingestion. The operator supplied filter rules and a mode.

Mode "whitelist": pass (output {"pass": true}) only if the post clearly satisfies at least one of the filter rules. Otherwise output {"pass": false}.

Mode "blacklist": pass ({"pass": true}) unless the post clearly matches a filter rule the operator wants excluded. If it matches an exclusion, output {"pass": false}.

Base your answer only on the post fields provided (title, body, subreddit, URL, author). If uncertain, prefer the conservative choice for the mode: whitelist → false, blacklist → true (allow).

Output valid JSON only with key "pass" (boolean), no other keys.`
	mode := "blacklist"
	if src.IsFilterWhitelist {
		mode = "whitelist"
	}
	var user strings.Builder
	fmt.Fprintf(&user, "Mode: %s\n\n", mode)
	fmt.Fprintf(&user, "Filter rules (lines or free text, apply according to the mode above):\n%s\n\n", strings.TrimSpace(src.Filter))
	fmt.Fprintf(&user, "Post subreddit: %s\n", strings.TrimSpace(post.Subreddit))
	fmt.Fprintf(&user, "Post author: %s\n", strings.TrimSpace(post.Author))
	fmt.Fprintf(&user, "Post is_self: %v\n", post.IsSelf)
	fmt.Fprintf(&user, "Post title:\n%s\n\n", strings.TrimSpace(post.Title))
	if u := strings.TrimSpace(post.URL); u != "" {
		fmt.Fprintf(&user, "Post URL: %s\n", u)
	}
	if p := strings.TrimSpace(post.Permalink); p != "" {
		fmt.Fprintf(&user, "Post permalink: %s\n", p)
	}
	fmt.Fprintf(&user, "Post selftext (body for text posts):\n%s\n", trimRunes(strings.TrimSpace(post.Selftext), redditFilterUserPromptMaxRunes))
	userStr := user.String()
	client, err := p_google_genai.NewClient(ctx)
	if err != nil {
		return false, "", err
	}
	temp := float32(0.1)
	var out redditFilterLLMResult
	cfg := &genai.GenerateContentConfig{
		SystemInstruction: genai.NewContentFromText(sys, genai.RoleUser),
		ResponseMIMEType:  "application/json",
		ResponseSchema:    p_google_genai.NewSchema[redditFilterLLMResult](),
		MaxOutputTokens:     redditFilterLlmMaxOutputTokens(),
		Temperature:       &temp,
	}
	resp, err := client.Models.GenerateContent(ctx, redditFilterLlmModel(), []*genai.Content{genai.NewContentFromText(userStr, genai.RoleUser)}, cfg)
	if err != nil {
		return false, userStr, err
	}
	if resp == nil {
		return false, userStr, fmt.Errorf("p_seer_reddit: nil filter model response")
	}
	raw := strings.TrimSpace(resp.Text())
	if raw == "" {
		return false, userStr, fmt.Errorf("p_seer_reddit: empty filter model response")
	}
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		return false, raw, fmt.Errorf("p_seer_reddit: filter model json: %w", err)
	}
	return out.Pass, raw, nil
}

// createRedditPostFilterRejectedTomb inserts a soft-deleted row with only [RedditPost.PostID] set
// (dedupe key) so the same id is not re-fetched and re-LLM'd.
func (r *RedditSource) createRedditPostFilterRejectedTomb(ctx context.Context, db *gorm.DB, pid string) error {
	now := time.Now().UTC()
	tomb := RedditPost{
		Model:  gorm.Model{DeletedAt: gorm.DeletedAt{Time: now, Valid: true}},
		PostID: pid,
	}
	err := db.WithContext(ctx).Create(&tomb).Error
	if err == nil {
		return nil
	}
	if isLikelyUniqueViolation(err) {
		return nil
	}
	return err
}
