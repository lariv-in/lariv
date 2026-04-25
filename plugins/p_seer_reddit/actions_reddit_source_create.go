package p_seer_reddit

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// RedditSourceCreateParams is the structured input for [CreateRedditSource]
// (same logical fields as the Reddit source create form).
type RedditSourceCreateParams struct {
	RedditRunnerID *uint
	Subreddits     []string
	SearchQuery         string
	Filter              string
	IsFilterWhitelist   bool
	MaxFreshPosts       uint
	LoadWebsites        bool
}

// ValidateRedditSourceCreate returns field-level errors keyed like form fields (e.g. Subreddits). Empty map means valid.
func ValidateRedditSourceCreate(p RedditSourceCreateParams) map[string]error {
	out := make(map[string]error)
	n := 0
	for _, s := range p.Subreddits {
		if strings.TrimSpace(s) != "" {
			n++
		}
	}
	if n == 0 {
		out["Subreddits"] = errors.New("add at least one subreddit")
	}
	return out
}

// CreateRedditSource inserts a new [RedditSource] after [ValidateRedditSourceCreate].
func CreateRedditSource(ctx context.Context, db *gorm.DB, p RedditSourceCreateParams) (RedditSource, error) {
	if db == nil {
		return RedditSource{}, fmt.Errorf("p_seer_reddit: CreateRedditSource: nil db")
	}
	if fe := ValidateRedditSourceCreate(p); len(fe) > 0 {
		return RedditSource{}, firstFieldError(fe)
	}
	maxFresh := p.MaxFreshPosts
	if maxFresh == 0 {
		maxFresh = defaultMaxFreshPosts
	}
	subs := make([]string, 0, len(p.Subreddits))
	for _, s := range p.Subreddits {
		if t := strings.TrimSpace(s); t != "" {
			subs = append(subs, t)
		}
	}
	b, err := json.Marshal(subs)
	if err != nil {
		return RedditSource{}, fmt.Errorf("p_seer_reddit: marshal subreddits: %w", err)
	}
	var runnerCopy *uint
	if p.RedditRunnerID != nil && *p.RedditRunnerID != 0 {
		v := *p.RedditRunnerID
		var n int64
		if err := db.WithContext(ctx).Model(&RedditRunner{}).Where("id = ?", v).Count(&n).Error; err != nil {
			return RedditSource{}, err
		}
		if n == 0 {
			return RedditSource{}, fmt.Errorf("p_seer_reddit: reddit runner id %d not found", v)
		}
		runnerCopy = &v
	}
	src := RedditSource{
		RedditRunnerID:      runnerCopy,
		Subreddits:          datatypes.JSON(b),
		SearchQuery:         strings.TrimSpace(p.SearchQuery),
		Filter:              strings.TrimSpace(p.Filter),
		IsFilterWhitelist:   p.IsFilterWhitelist,
		MaxFreshPosts:       maxFresh,
		LoadWebsites:        p.LoadWebsites,
	}
	if err := db.WithContext(ctx).Create(&src).Error; err != nil {
		return RedditSource{}, err
	}
	return src, nil
}

func firstFieldError(fe map[string]error) error {
	for _, e := range fe {
		if e != nil {
			return e
		}
	}
	return errors.New("validation failed")
}

// RedditSourceCreateParamsFromFormMap parses the create form map into [RedditSourceCreateParams].
func RedditSourceCreateParamsFromFormMap(formData map[string]any) (RedditSourceCreateParams, error) {
	var p RedditSourceCreateParams
	subRaw, ok := formData["Subreddits"]
	if !ok {
		return p, errors.New("missing subreddits")
	}
	var b []byte
	switch v := subRaw.(type) {
	case datatypes.JSON:
		b = []byte(v)
	case []byte:
		b = v
	case string:
		b = []byte(v)
	default:
		return p, errors.New("invalid subreddits value")
	}
	var subs []string
	if err := json.Unmarshal(b, &subs); err != nil {
		return p, err
	}
	p.Subreddits = subs
	if rid, ok := redditRunnerIDFromFormMap(formData); ok {
		p.RedditRunnerID = rid
	}
	if sq, ok := formData["SearchQuery"].(string); ok {
		p.SearchQuery = sq
	}
	if f, ok := formData["Filter"].(string); ok {
		p.Filter = f
	}
	if w, ok := formData["IsFilterWhitelist"].(bool); ok {
		p.IsFilterWhitelist = w
	}
	p.MaxFreshPosts = uintFromFormAny(formData["MaxFreshPosts"])
	if lw, ok := formData["LoadWebsites"].(bool); ok {
		p.LoadWebsites = lw
	}
	return p, nil
}

func uintFromFormAny(v any) uint {
	switch x := v.(type) {
	case uint:
		return x
	case int:
		if x > 0 {
			return uint(x)
		}
	case int64:
		if x > 0 {
			return uint(x)
		}
	case float64:
		if x > 0 {
			return uint(x)
		}
	case json.Number:
		n, err := x.Int64()
		if err == nil && n > 0 {
			return uint(n)
		}
	}
	return 0
}
