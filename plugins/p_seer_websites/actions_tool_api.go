package p_seer_websites

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/lariv-in/lago/lago"
	"gorm.io/gorm"
)

// WebsiteSourceCreateParams is API-style input for creating a [WebsiteSource] (assistant / automation).
type WebsiteSourceCreateParams struct {
	SeedURL         string
	Depth           uint
	WebsiteRunnerID *uint
}

// WebsiteSourceUpdateParams replaces editable fields on a [WebsiteSource].
type WebsiteSourceUpdateParams struct {
	SourceID        uint
	SeedURL         string
	Depth           uint
	WebsiteRunnerID *uint
}

func pageURLFromValidatedSeed(ctx context.Context, raw string) (lago.PageURL, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return lago.PageURL{}, errors.New("seed url is required")
	}
	u, err := fetchableWebsiteURL(ctx, raw)
	if err != nil {
		return lago.PageURL{}, err
	}
	var pp lago.PageURL
	pp.SetFromURL(u)
	return pp, nil
}

func clampWebsiteSourceDepth(d uint) uint {
	if d > maxWebsiteSourceDepth {
		return maxWebsiteSourceDepth
	}
	return d
}

func verifyWebsiteRunnerFK(ctx context.Context, db *gorm.DB, id *uint) (*uint, error) {
	if id == nil || *id == 0 {
		return nil, nil
	}
	v := *id
	var n int64
	if err := db.WithContext(ctx).Model(&WebsiteRunner{}).Where("id = ?", v).Count(&n).Error; err != nil {
		return nil, err
	}
	if n == 0 {
		return nil, fmt.Errorf("p_seer_websites: website runner id %d not found", v)
	}
	return &v, nil
}

// CreateWebsiteSourceFromParams inserts a [WebsiteSource] after URL + depth validation.
func CreateWebsiteSourceFromParams(ctx context.Context, db *gorm.DB, p WebsiteSourceCreateParams) (WebsiteSource, error) {
	if db == nil {
		return WebsiteSource{}, fmt.Errorf("p_seer_websites: CreateWebsiteSourceFromParams: nil db")
	}
	pp, err := pageURLFromValidatedSeed(ctx, p.SeedURL)
	if err != nil {
		return WebsiteSource{}, err
	}
	runner, err := verifyWebsiteRunnerFK(ctx, db, p.WebsiteRunnerID)
	if err != nil {
		return WebsiteSource{}, err
	}
	d := clampWebsiteSourceDepth(p.Depth)
	src := WebsiteSource{
		WebsiteRunnerID: runner,
		URL:             pp,
		Depth:           d,
	}
	if err := db.WithContext(ctx).Create(&src).Error; err != nil {
		return WebsiteSource{}, err
	}
	return src, nil
}

// UpdateWebsiteSourceFromParams updates an existing [WebsiteSource].
func UpdateWebsiteSourceFromParams(ctx context.Context, db *gorm.DB, p WebsiteSourceUpdateParams) (WebsiteSource, error) {
	if db == nil {
		return WebsiteSource{}, fmt.Errorf("p_seer_websites: UpdateWebsiteSourceFromParams: nil db")
	}
	if p.SourceID == 0 {
		return WebsiteSource{}, fmt.Errorf("p_seer_websites: website_source_id required")
	}
	var existing WebsiteSource
	if err := db.WithContext(ctx).Where("id = ?", p.SourceID).Take(&existing).Error; err != nil {
		return WebsiteSource{}, fmt.Errorf("p_seer_websites: load website source: %w", err)
	}
	pp, err := pageURLFromValidatedSeed(ctx, p.SeedURL)
	if err != nil {
		return WebsiteSource{}, err
	}
	runner, err := verifyWebsiteRunnerFK(ctx, db, p.WebsiteRunnerID)
	if err != nil {
		return WebsiteSource{}, err
	}
	d := clampWebsiteSourceDepth(p.Depth)
	updates := map[string]any{
		"url":               pp,
		"depth":             d,
		"website_runner_id": runner,
	}
	if err := db.WithContext(ctx).Model(&WebsiteSource{}).Where("id = ?", p.SourceID).Updates(updates).Error; err != nil {
		return WebsiteSource{}, err
	}
	var out WebsiteSource
	if err := db.WithContext(ctx).Where("id = ?", p.SourceID).Take(&out).Error; err != nil {
		return WebsiteSource{}, err
	}
	return out, nil
}

// WebsiteRunnerCreateParams creates a [WebsiteRunner] by name and Go duration string is parsed by caller.
type WebsiteRunnerCreateParams struct {
	Name     string
	Duration time.Duration
}

// CreateWebsiteRunnerFromParams inserts a runner after validation.
func CreateWebsiteRunnerFromParams(ctx context.Context, db *gorm.DB, p WebsiteRunnerCreateParams) (WebsiteRunner, error) {
	if db == nil {
		return WebsiteRunner{}, fmt.Errorf("p_seer_websites: CreateWebsiteRunnerFromParams: nil db")
	}
	name := strings.TrimSpace(p.Name)
	if name == "" {
		return WebsiteRunner{}, errors.New("p_seer_websites: worker name required")
	}
	if p.Duration <= 0 {
		return WebsiteRunner{}, errors.New("p_seer_websites: duration must be positive")
	}
	var n int64
	if err := db.WithContext(ctx).Model(&WebsiteRunner{}).Where("name = ?", name).Count(&n).Error; err != nil {
		return WebsiteRunner{}, err
	}
	if n > 0 {
		return WebsiteRunner{}, fmt.Errorf("p_seer_websites: worker name %q already exists", name)
	}
	w := WebsiteRunner{Name: name, Duration: p.Duration}
	if err := db.WithContext(ctx).Create(&w).Error; err != nil {
		return WebsiteRunner{}, err
	}
	return w, nil
}

// WebsiteRunnerUpdateParams replaces name and duration on an existing [WebsiteRunner].
type WebsiteRunnerUpdateParams struct {
	ID       uint
	Name     string
	Duration time.Duration
}

// UpdateWebsiteRunnerFromParams persists name + duration (unique name enforced).
func UpdateWebsiteRunnerFromParams(ctx context.Context, db *gorm.DB, p WebsiteRunnerUpdateParams) (WebsiteRunner, error) {
	if db == nil {
		return WebsiteRunner{}, fmt.Errorf("p_seer_websites: UpdateWebsiteRunnerFromParams: nil db")
	}
	if p.ID == 0 {
		return WebsiteRunner{}, errors.New("p_seer_websites: website_runner_id required")
	}
	name := strings.TrimSpace(p.Name)
	if name == "" {
		return WebsiteRunner{}, errors.New("p_seer_websites: worker name required")
	}
	if p.Duration <= 0 {
		return WebsiteRunner{}, errors.New("p_seer_websites: duration must be positive")
	}
	var existing WebsiteRunner
	if err := db.WithContext(ctx).Where("id = ?", p.ID).Take(&existing).Error; err != nil {
		return WebsiteRunner{}, fmt.Errorf("p_seer_websites: load website runner: %w", err)
	}
	var conflict int64
	if err := db.WithContext(ctx).Model(&WebsiteRunner{}).
		Where("name = ? AND id <> ?", name, p.ID).
		Count(&conflict).Error; err != nil {
		return WebsiteRunner{}, err
	}
	if conflict > 0 {
		return WebsiteRunner{}, fmt.Errorf("p_seer_websites: worker name %q already exists", name)
	}
	if err := db.WithContext(ctx).Model(&WebsiteRunner{}).Where("id = ?", p.ID).
		Updates(map[string]any{"name": name, "duration": p.Duration}).Error; err != nil {
		return WebsiteRunner{}, err
	}
	var out WebsiteRunner
	if err := db.WithContext(ctx).Where("id = ?", p.ID).Take(&out).Error; err != nil {
		return WebsiteRunner{}, err
	}
	return out, nil
}

type websiteSourceListRow struct {
	ID              uint   `json:"id"`
	SeedURL         string `json:"seed_url"`
	Depth           uint   `json:"depth"`
	WebsiteRunnerID *uint  `json:"website_runner_id,omitempty"`
}

type websiteRunnerListRow struct {
	ID       uint   `json:"id"`
	Name     string `json:"name"`
	Duration string `json:"duration"`
}

// ListWebsiteSourcesJSON returns non-deleted sources as JSON for tools.
func ListWebsiteSourcesJSON(ctx context.Context, db *gorm.DB) (string, error) {
	if db == nil {
		return "", fmt.Errorf("p_seer_websites: ListWebsiteSourcesJSON: nil db")
	}
	var rows []WebsiteSource
	if err := db.WithContext(ctx).Where("deleted_at IS NULL").Order("id DESC").Find(&rows).Error; err != nil {
		return "", err
	}
	out := make([]websiteSourceListRow, 0, len(rows))
	for _, s := range rows {
		var rid *uint
		if s.WebsiteRunnerID != nil && *s.WebsiteRunnerID != 0 {
			v := *s.WebsiteRunnerID
			rid = &v
		}
		out = append(out, websiteSourceListRow{
			ID:              s.ID,
			SeedURL:         strings.TrimSpace(s.URL.String()),
			Depth:           s.Depth,
			WebsiteRunnerID: rid,
		})
	}
	b, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// ListWebsiteRunnersJSON returns runners as JSON for tools.
func ListWebsiteRunnersJSON(ctx context.Context, db *gorm.DB) (string, error) {
	if db == nil {
		return "", fmt.Errorf("p_seer_websites: ListWebsiteRunnersJSON: nil db")
	}
	var rows []WebsiteRunner
	if err := db.WithContext(ctx).Where("deleted_at IS NULL").Order("id DESC").Find(&rows).Error; err != nil {
		return "", err
	}
	out := make([]websiteRunnerListRow, 0, len(rows))
	for _, w := range rows {
		out = append(out, websiteRunnerListRow{
			ID:       w.ID,
			Name:     w.Name,
			Duration: w.Duration.String(),
		})
	}
	b, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		return "", err
	}
	return string(b), nil
}
