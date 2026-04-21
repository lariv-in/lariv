package p_seer_websites

import (
	"context"
	"fmt"
	"log/slog"
	"sync/atomic"
	"time"

	"github.com/lariv-in/lago/plugins/p_seer_intel"
	"gorm.io/gorm"
)

// websiteIntelIngestActive limits concurrent website→Intel jobs (separate from Reddit gate).
var websiteIntelIngestActive atomic.Bool

func createIntelForWebsiteIfMissing(ctx context.Context, db *gorm.DB, site Website) error {
	kind := (Website{}).Kind()
	exists, err := p_seer_intel.IntelExistsForSource(ctx, db, kind, site.ID)
	if err != nil {
		return fmt.Errorf("exists check: %w", err)
	}
	if exists {
		return nil
	}
	intel, err := p_seer_intel.NewFromIntelKind(ctx, &site)
	if err != nil {
		return fmt.Errorf("generate: %w", err)
	}
	if err := p_seer_intel.CreateIntelAndEvent(ctx, db, &intel); err != nil {
		return fmt.Errorf("persist: %w", err)
	}
	return nil
}

// RunWebsiteSingleIntelIngest runs [createIntelForWebsiteIfMissing] for one row in the background window.
func RunWebsiteSingleIntelIngest(ctx context.Context, db *gorm.DB, site Website) {
	ctx, cancel := context.WithTimeout(ctx, 15*time.Minute)
	defer cancel()
	if err := createIntelForWebsiteIfMissing(ctx, db, site); err != nil {
		slog.Warn("p_seer_websites: single add intel", "website_id", site.ID, "error", err)
	}
}

// RunWebsitesBulkIntelIngest runs [createIntelForWebsiteIfMissing] for each site (skips rows that already have Intel).
func RunWebsitesBulkIntelIngest(ctx context.Context, db *gorm.DB, sites []Website) {
	for _, site := range sites {
		if err := createIntelForWebsiteIfMissing(ctx, db, site); err != nil {
			slog.Warn("p_seer_websites: bulk add intel", "website_id", site.ID, "error", err)
		}
	}
}
