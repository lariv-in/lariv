package p_seer_deepsearch

import (
	"context"
	"log/slog"
	"strings"
	"unicode/utf8"

	"gorm.io/gorm"
)

const deepSearchLogMessageMaxRunes = 60000

// appendDeepSearchLog inserts one [DeepSearchLog] row. Truncates very long messages.
func appendDeepSearchLog(ctx context.Context, db *gorm.DB, deepSearchID uint, kind, message string) {
	if db == nil || deepSearchID == 0 {
		return
	}
	kind = strings.TrimSpace(kind)
	if kind == "" {
		kind = DeepSearchLogKindInfo
	}
	msg := strings.TrimSpace(message)
	if utf8.RuneCountInString(msg) > deepSearchLogMessageMaxRunes {
		runes := []rune(msg)
		msg = string(runes[:deepSearchLogMessageMaxRunes]) + "\n…(truncated)"
	}
	row := DeepSearchLog{
		DeepSearchID: deepSearchID,
		Kind:         kind,
		Message:      msg,
	}
	if err := db.WithContext(ctx).Create(&row).Error; err != nil {
		slog.Error("p_seer_deepsearch: append log", "deep_search_id", deepSearchID, "kind", kind, "error", err)
	}
}
