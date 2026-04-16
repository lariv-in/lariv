package p_lacerate

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/lariv-in/lago/lago"
	"gorm.io/gorm"
)

const (
	sourceKindGoogleSearch = "google_search"
)

type GoogleSearchSource struct {
	gorm.Model
	Query    string `gorm:"type:text;not null;default:''"`
	SourceID uint   `gorm:"not null;uniqueIndex"`
	Source   Source `gorm:"foreignKey:SourceID"`
}

func (g GoogleSearchSource) Fetch(ctx context.Context, db *gorm.DB, existingDedup map[string]struct{}) ([]Intel, error) {
	query := strings.TrimSpace(g.Query)
	if query == "" {
		err := fmt.Errorf("google search source query is empty")
		slog.Error("lacerate: google search source fetch", "error", err, "source_id", g.SourceID)
		return nil, err
	}

	sourceID := g.SourceID
	out, err := runWebsearchQueryFetch(ctx, db, &sourceID, query, existingDedup)
	if err != nil {
		slog.Error("lacerate: google search source query", "error", err, "source_id", g.SourceID, "query", query)
		return nil, err
	}
	return out, nil
}

func init() {
	SourceKindMap[sourceKindGoogleSearch] = SourceDesc{
		Name:  "GoogleSearch",
		Model: GoogleSearchSource{},
	}
	if err := RegistrySourceKind.Register(sourceKindGoogleSearch, func() SourceInterface { return &GoogleSearchSource{} }); err != nil {
		panic(err)
	}
	lago.OnDBInit("p_lacerate.google_search_source_model", func(db *gorm.DB) *gorm.DB {
		lago.RegisterModel[GoogleSearchSource](db)
		return db
	})
}
