package p_seer_reddit

import (
	"context"

	"github.com/lariv-in/lago/plugins/p_seer_intel"
	"gorm.io/gorm"
)

func init() {
	_ = p_seer_intel.RegistryIntelKind.Register((RedditPost{}).Kind(), loadRedditPostIntelKind)
}

func loadRedditPostIntelKind(ctx context.Context, db *gorm.DB, id uint) (p_seer_intel.IntelKind, error) {
	var rp RedditPost
	if err := db.WithContext(ctx).First(&rp, id).Error; err != nil {
		return nil, err
	}
	return new(rp), nil
}
