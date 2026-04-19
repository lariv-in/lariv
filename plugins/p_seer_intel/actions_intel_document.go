package p_seer_intel

import (
	"context"
	"fmt"
	"strings"

	"gorm.io/gorm"
)

// IntelToolDocument bundles display fields plus full source text for LLM tools.
type IntelToolDocument struct {
	ID      uint
	Title   string
	Summary string
	Kind    string
	KindID  uint
	Content string
}

// IntelDocumentForTool loads one [Intel] and the backing [IntelKind] content.
func IntelDocumentForTool(ctx context.Context, db *gorm.DB, intelID uint) (IntelToolDocument, error) {
	var out IntelToolDocument
	if db == nil {
		return out, fmt.Errorf("p_seer_intel: IntelDocumentForTool: db is nil")
	}
	if intelID == 0 {
		return out, fmt.Errorf("p_seer_intel: IntelDocumentForTool: intel id is zero")
	}

	var row Intel
	if err := db.WithContext(ctx).First(&row, intelID).Error; err != nil {
		return out, err
	}
	out.ID = row.ID
	out.Title = row.Title
	out.Summary = row.Summary
	out.Kind = strings.TrimSpace(row.Kind)
	out.KindID = row.KindID

	if out.Kind == "" || out.KindID == 0 {
		return out, fmt.Errorf("p_seer_intel: IntelDocumentForTool: intel %d missing kind/kind_id", intelID)
	}
	k, err := LoadIntelKind(ctx, db, out.Kind, out.KindID)
	if err != nil {
		return out, err
	}
	if k == nil {
		return out, fmt.Errorf("p_seer_intel: IntelDocumentForTool: nil IntelKind for intel %d", intelID)
	}
	out.Content = strings.TrimSpace(k.Content())
	return out, nil
}
