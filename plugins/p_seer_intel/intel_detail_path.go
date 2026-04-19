package p_seer_intel

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	"gorm.io/gorm"
)

// IntelDetailPathForSource returns the app path to [DetailRoute] for the [Intel] row with the given Kind and KindID.
func IntelDetailPathForSource(ctx context.Context, kind string, kindID uint) (string, error) {
	if kind == "" || kindID == 0 {
		return "", fmt.Errorf("p_seer_intel: IntelDetailPathForSource: empty kind or kind id")
	}
	db, err := getters.DBFromContext(ctx)
	if err != nil {
		return "", err
	}
	var row Intel
	if err := db.WithContext(ctx).Where("kind = ? AND kind_id = ?", kind, kindID).First(&row).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", err
		}
		return "", err
	}
	return lago.RoutePath("seer_intel.DetailRoute", map[string]getters.Getter[any]{
		"id": getters.Any(getters.Static(strconv.FormatUint(uint64(row.ID), 10))),
	})(ctx)
}
