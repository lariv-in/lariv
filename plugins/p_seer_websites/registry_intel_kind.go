package p_seer_websites

import (
	"context"

	"github.com/lariv-in/lago/plugins/p_seer_intel"
	"gorm.io/gorm"
)

func init() {
	_ = p_seer_intel.RegistryIntelKind.Register((Website{}).Kind(), loadWebsiteIntelKind)
}

func loadWebsiteIntelKind(ctx context.Context, db *gorm.DB, id uint) (p_seer_intel.IntelKind, error) {
	var w Website
	if err := db.WithContext(ctx).First(&w, id).Error; err != nil {
		return nil, err
	}
	return new(w), nil
}
