package p_seer_reddit

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// RedditRunnerCreateParams creates a [RedditRunner] by name and cadence duration.
type RedditRunnerCreateParams struct {
	Name     string
	Duration time.Duration
}

// CreateRedditRunner inserts a runner after validation.
func CreateRedditRunner(ctx context.Context, db *gorm.DB, p RedditRunnerCreateParams) (RedditRunner, error) {
	if db == nil {
		return RedditRunner{}, fmt.Errorf("p_seer_reddit: CreateRedditRunner: nil db")
	}
	name := strings.TrimSpace(p.Name)
	if name == "" {
		return RedditRunner{}, fmt.Errorf("p_seer_reddit: CreateRedditRunner: empty name")
	}
	if p.Duration <= 0 {
		return RedditRunner{}, fmt.Errorf("p_seer_reddit: CreateRedditRunner: duration must be positive")
	}
	var conflict int64
	if err := db.WithContext(ctx).Model(&RedditRunner{}).Where("name = ?", name).Count(&conflict).Error; err != nil {
		return RedditRunner{}, err
	}
	if conflict > 0 {
		return RedditRunner{}, fmt.Errorf("p_seer_reddit: runner name %q already exists", name)
	}
	runner := RedditRunner{Name: name, Duration: p.Duration}
	if err := db.WithContext(ctx).Create(&runner).Error; err != nil {
		return RedditRunner{}, err
	}
	return runner, nil
}

// RedditSourceUpdateParams replaces editable fields on an existing [RedditSource].
// Fields match [RedditSourceCreateParams]; [SourceID] targets the row.
type RedditSourceUpdateParams struct {
	SourceID uint
	RedditSourceCreateParams
}

// UpdateRedditSource persists a full replace of editable columns after validation.
func UpdateRedditSource(ctx context.Context, db *gorm.DB, p RedditSourceUpdateParams) (RedditSource, error) {
	if db == nil {
		return RedditSource{}, fmt.Errorf("p_seer_reddit: UpdateRedditSource: nil db")
	}
	if p.SourceID == 0 {
		return RedditSource{}, fmt.Errorf("p_seer_reddit: UpdateRedditSource: source id required")
	}
	var existing RedditSource
	if err := db.WithContext(ctx).Where("id = ?", p.SourceID).Take(&existing).Error; err != nil {
		return RedditSource{}, fmt.Errorf("p_seer_reddit: UpdateRedditSource: %w", err)
	}
	if fe := ValidateRedditSourceCreate(p.RedditSourceCreateParams); len(fe) > 0 {
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
	updates := map[string]any{
		"reddit_runner_id":      runnerCopy,
		"subreddits":            datatypes.JSON(b),
		"search_query":          strings.TrimSpace(p.SearchQuery),
		"filter":                strings.TrimSpace(p.Filter),
		"is_filter_whitelist":   p.IsFilterWhitelist,
		"max_fresh_posts":       maxFresh,
		"load_websites":         p.LoadWebsites,
	}
	if err := db.WithContext(ctx).Model(&RedditSource{}).Where("id = ?", p.SourceID).Updates(updates).Error; err != nil {
		return RedditSource{}, err
	}
	var out RedditSource
	if err := db.WithContext(ctx).Where("id = ?", p.SourceID).Take(&out).Error; err != nil {
		return RedditSource{}, err
	}
	return out, nil
}

// RedditRunnerUpdateParams replaces [RedditRunner] name and cadence duration.
type RedditRunnerUpdateParams struct {
	ID       uint
	Name     string
	Duration time.Duration
}

// UpdateRedditRunner persists name + duration after validation.
func UpdateRedditRunner(ctx context.Context, db *gorm.DB, p RedditRunnerUpdateParams) (RedditRunner, error) {
	if db == nil {
		return RedditRunner{}, fmt.Errorf("p_seer_reddit: UpdateRedditRunner: nil db")
	}
	if p.ID == 0 {
		return RedditRunner{}, fmt.Errorf("p_seer_reddit: UpdateRedditRunner: runner id required")
	}
	name := strings.TrimSpace(p.Name)
	if name == "" {
		return RedditRunner{}, fmt.Errorf("p_seer_reddit: UpdateRedditRunner: empty name")
	}
	if p.Duration <= 0 {
		return RedditRunner{}, fmt.Errorf("p_seer_reddit: UpdateRedditRunner: duration must be positive")
	}
	var existing RedditRunner
	if err := db.WithContext(ctx).Where("id = ?", p.ID).Take(&existing).Error; err != nil {
		return RedditRunner{}, fmt.Errorf("p_seer_reddit: UpdateRedditRunner: %w", err)
	}
	var conflict int64
	if err := db.WithContext(ctx).Model(&RedditRunner{}).
		Where("name = ? AND id <> ?", name, p.ID).
		Count(&conflict).Error; err != nil {
		return RedditRunner{}, err
	}
	if conflict > 0 {
		return RedditRunner{}, fmt.Errorf("p_seer_reddit: runner name %q already exists", name)
	}
	if err := db.WithContext(ctx).Model(&RedditRunner{}).Where("id = ?", p.ID).
		Updates(map[string]any{"name": name, "duration": p.Duration}).Error; err != nil {
		return RedditRunner{}, err
	}
	var out RedditRunner
	if err := db.WithContext(ctx).Where("id = ?", p.ID).Take(&out).Error; err != nil {
		return RedditRunner{}, err
	}
	return out, nil
}
