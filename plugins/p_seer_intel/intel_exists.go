package p_seer_intel

import (
	"context"

	"gorm.io/gorm"
)

// IntelExistsForSource reports whether an [Intel] row already exists for the given kind and source row id.
func IntelExistsForSource(ctx context.Context, db *gorm.DB, kind string, kindID uint) (bool, error) {
	var n int64
	if err := db.WithContext(ctx).Model(&Intel{}).Where("kind = ? AND kind_id = ?", kind, kindID).Count(&n).Error; err != nil {
		return false, err
	}
	return n > 0, nil
}
