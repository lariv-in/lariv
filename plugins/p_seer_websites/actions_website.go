package p_seer_websites

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	"gorm.io/gorm"
)

// ErrWebsiteSoftDeleteNotFound is returned when no active row matches the id for [SoftWipeWebsite].
var ErrWebsiteSoftDeleteNotFound = errors.New("p_seer_websites: website not found or already deleted")

// SoftWipeWebsite sets [gorm.Model.DeletedAt] and clears [Website.Markdown]; keeps [Website.URL] for audit/dedup.
func SoftWipeWebsite(ctx context.Context, db *gorm.DB, id uint) error {
	if id == 0 {
		return fmt.Errorf("p_seer_websites: invalid website id")
	}
	now := time.Now().UTC()
	tx := db.WithContext(ctx).Model(&Website{}).
		Where("id = ? AND deleted_at IS NULL", id).
		Select("DeletedAt", "Markdown").
		Updates(&Website{
			Model:    gorm.Model{DeletedAt: gorm.DeletedAt{Time: now, Valid: true}},
			Markdown: "",
		})
	if tx.Error != nil {
		return tx.Error
	}
	if tx.RowsAffected == 0 {
		return ErrWebsiteSoftDeleteNotFound
	}
	return nil
}

// Kind satisfies [p_seer_intel.IntelKind] for [Website].
func (Website) Kind() string {
	return "website"
}

// IntelID satisfies [p_seer_intel.IntelKind] for [Website].
func (w Website) IntelID() uint {
	return w.ID
}

// Content satisfies [p_seer_intel.IntelKind] for [*Website].
func (w *Website) Content() string {
	if w == nil {
		return ""
	}
	var b strings.Builder
	if u := w.URL.URLPtr(); u != nil {
		b.WriteString("# ")
		b.WriteString(websiteTitleHint(u))
		b.WriteString("\n\n")
	}
	if md := strings.TrimSpace(w.Markdown); md != "" {
		b.WriteString(md)
		b.WriteString("\n\n")
	}
	b.WriteString("---\n\n")
	if u := w.URL.URLPtr(); u != nil {
		fmt.Fprintf(&b, "- **URL:** %s\n", u.String())
	}
	return strings.TrimSpace(b.String())
}

// IntelDetail satisfies [p_seer_intel.IntelKind] for [*Website].
func (w *Website) IntelDetail(ctx context.Context) (string, error) {
	if w == nil || w.ID == 0 {
		return "", fmt.Errorf("p_seer_websites: IntelDetail: missing website")
	}
	return lago.RoutePath("seer_websites.WebsiteDetailRoute", map[string]getters.Getter[any]{
		"id": getters.Any(getters.Static(strconv.FormatUint(uint64(w.ID), 10))),
	})(ctx)
}
