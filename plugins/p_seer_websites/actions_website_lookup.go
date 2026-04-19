package p_seer_websites

import (
	"context"
	"errors"
	"fmt"
	"net/url"

	"gorm.io/gorm"
)

// WebsiteByFetchableURL returns the active [Website] row for the canonical URL produced by
// [fetchableWebsiteURL], matching the key used in [WebsiteScrapeIfAbsent].
func WebsiteByFetchableURL(ctx context.Context, db *gorm.DB, u *url.URL) (Website, error) {
	var zero Website
	if u == nil {
		return zero, errors.New("p_seer_websites: WebsiteByFetchableURL: url is nil")
	}
	if db == nil {
		return zero, errors.New("p_seer_websites: WebsiteByFetchableURL: db is nil")
	}
	canon, err := fetchableWebsiteURL(ctx, u.String())
	if err != nil {
		return zero, err
	}
	key := canon.String()

	var w Website
	if err := db.WithContext(ctx).Where("url = ? AND deleted_at IS NULL", key).First(&w).Error; err != nil {
		return zero, err
	}
	return w, nil
}

// WebsiteIDByFetchableURL returns the primary key for [WebsiteByFetchableURL] or an error
// if no active row exists.
func WebsiteIDByFetchableURL(ctx context.Context, db *gorm.DB, u *url.URL) (uint, error) {
	w, err := WebsiteByFetchableURL(ctx, db, u)
	if err != nil {
		return 0, err
	}
	if w.ID == 0 {
		return 0, fmt.Errorf("p_seer_websites: WebsiteIDByFetchableURL: zero id for %s", u.String())
	}
	return w.ID, nil
}
