package p_lacerate

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/lariv-in/lago/lago"
	"gorm.io/gorm"
)

const sourceKindWebsearch = "websearch"

type WebsearchSource struct {
	gorm.Model
	Query    string `gorm:"type:text;not null;default:''"`
	SourceID uint   `gorm:"not null;uniqueIndex"`
	Source   Source `gorm:"foreignKey:SourceID"`
}

func (w WebsearchSource) Fetch(ctx context.Context, db *gorm.DB, existingDedup map[string]struct{}) ([]Intel, error) {
	query := strings.TrimSpace(w.Query)
	if query == "" {
		err := fmt.Errorf("websearch source query is empty")
		slog.Error("lacerate: websearch source fetch", "error", err, "source_id", w.SourceID)
		return nil, err
	}
	sourceID := w.SourceID
	intels, err := runWebsearchQueryFetch(ctx, db, &sourceID, query, existingDedup)
	if err != nil {
		slog.Error("lacerate: websearch source query", "error", err, "source_id", w.SourceID, "query", query)
		return nil, err
	}
	return intels, nil
}

type Websearch struct {
	gorm.Model
	Query         string `gorm:"type:text;not null;default:''"`
	Status        string `gorm:"type:text;not null;default:'queued'"`
	LastRunError  string `gorm:"type:text;not null;default:''"`
	LastStartedAt *time.Time
	LastEndedAt   *time.Time
	IntelRows     []WebsearchIntel `gorm:"foreignKey:WebsearchID"`
}

type WebsearchIntel struct {
	gorm.Model
	WebsearchID uint      `gorm:"not null;index"`
	Websearch   Websearch `gorm:"foreignKey:WebsearchID;constraint:OnDelete:CASCADE"`
	IntelID     uint      `gorm:"not null;index"`
	Intel       Intel     `gorm:"foreignKey:IntelID;constraint:OnDelete:CASCADE"`
}

func (Websearch) TableName() string {
	return "websearches"
}

func init() {
	SourceKindMap[sourceKindWebsearch] = SourceDesc{
		Name:  "Websearch",
		Model: WebsearchSource{},
	}
	if err := RegistrySourceKind.Register(sourceKindWebsearch, func() SourceInterface { return &WebsearchSource{} }); err != nil {
		panic(err)
	}
	lago.OnDBInit("p_lacerate.websearch_source_model", func(db *gorm.DB) *gorm.DB {
		lago.RegisterModel[WebsearchSource](db)
		lago.RegisterModel[Websearch](db)
		lago.RegisterModel[WebsearchIntel](db)
		return db
	})
}
