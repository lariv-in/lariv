package p_lacerate

import (
	"context"

	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/registry"
	"github.com/pgvector/pgvector-go"
	"gorm.io/gorm"
)

// TargetOfInterestTypeChoices is the persisted value (Key) and UI label (Value) for [TargetOfInterest.Type].
var TargetOfInterestTypeChoices = []registry.Pair[string, string]{
	{Key: "report", Value: "Report"},
	{Key: "briefing", Value: "Briefing"},
	{Key: "memo", Value: "Memo"},
	{Key: "dataset", Value: "Dataset summary"},
	{Key: "other", Value: "Other"},
}

// TargetOfInterest is a Target of Interest: a manually curated document (reports, etc.) with an optional embedding for retrieval when [GeminiEmbeddingConfig] / [VLEmbedder] is configured.
type TargetOfInterest struct {
	gorm.Model
	Name        string `gorm:"not null"`
	Description string `gorm:"type:text"`
	// Type is a key from [TargetOfInterestTypeChoices].
	Type    string `gorm:"not null"`
	Content string `gorm:"type:text;not null;default:''"`
	// Embedding matches [IntelEmbeddingDim] and the configured [VLEmbedder].
	Embedding *pgvector.Vector `gorm:"type:vector(1024)"`
}

func (TargetOfInterest) TableName() string { return "targets_of_interest" }

// BeforeSave sets [TargetOfInterest.Embedding] from name/type/description/content ([prepareTargetOfInterestEmbeddingForSave]).
func (a *TargetOfInterest) BeforeSave(tx *gorm.DB) error {
	if tx.Statement.SkipHooks {
		return nil
	}
	prepareTargetOfInterestEmbeddingForSave(context.Background(), a)
	return nil
}

func init() {
	lago.OnDBInit("p_lacerate.target_of_interest_model", func(db *gorm.DB) *gorm.DB {
		lago.RegisterModel[TargetOfInterest](db)
		return db
	})
}
