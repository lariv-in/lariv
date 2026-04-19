package p_seer_websites

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/url"

	"github.com/lariv-in/lago/lago"
	"gorm.io/gorm"
)

const websiteScrapeURLQueueCap = 64

// WebsiteScrapeURLQueue carries [*url.URL] jobs for [runWebsiteScrapeURLQueueWorker].
// Send only non-nil URLs; nil is ignored. The worker starts after DB init ([lago.OnDBInit]).
var WebsiteScrapeURLQueue = make(chan *url.URL, websiteScrapeURLQueueCap)

func init() {
	lago.OnDBInit("p_seer_websites.scrape_url_worker", func(db *gorm.DB) *gorm.DB {
		ctx := context.Background()
		go runWebsiteScrapeURLQueueWorker(ctx, db)
		return db
	})
}

func runWebsiteScrapeURLQueueWorker(ctx context.Context, db *gorm.DB) {
	slog.Info("p_seer_websites: scrape URL queue worker started")
	for u := range WebsiteScrapeURLQueue {
		if u == nil {
			continue
		}
		if err := WebsiteScrapeIfAbsent(ctx, db, u); err != nil {
			slog.Warn("p_seer_websites: queue scrape", "url", u.String(), "error", err)
		}
	}
}

// WebsiteScrapeIfAbsent runs the same scrape path as the add form when no active [Website]
// row already stores that canonical URL (string match on column url, non-deleted rows only).
// Concurrent sends for the same URL can still race; only one row typically wins.
func WebsiteScrapeIfAbsent(ctx context.Context, db *gorm.DB, u *url.URL) error {
	if u == nil {
		return errors.New("p_seer_websites: WebsiteScrapeIfAbsent: url is nil")
	}
	if db == nil {
		return errors.New("p_seer_websites: WebsiteScrapeIfAbsent: db is nil")
	}
	canon, err := fetchableWebsiteURL(ctx, u.String())
	if err != nil {
		return err
	}
	key := canon.String()

	var n int64
	if err := db.WithContext(ctx).Model(&Website{}).
		Where("url = ? AND deleted_at IS NULL", key).
		Count(&n).Error; err != nil {
		return fmt.Errorf("exists check: %w", err)
	}
	if n > 0 {
		return nil
	}

	md, outCanon, err := ScrapeToMarkdown(ctx, key)
	if err != nil {
		return err
	}

	var pp lago.PageURL
	pp.SetFromURL(outCanon)
	w := Website{URL: pp, Markdown: md}
	if err := db.WithContext(ctx).Create(&w).Error; err != nil {
		return fmt.Errorf("create website: %w", err)
	}
	return nil
}
