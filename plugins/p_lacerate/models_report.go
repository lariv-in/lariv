package p_lacerate

import (
	"context"
	"fmt"
	"strings"

	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/registry"
	"github.com/pgvector/pgvector-go"
	"gorm.io/gorm"
)

// ReportTypeChoices is the persisted value (Key) and UI label (Value) for [Report.Type].
var ReportTypeChoices = []registry.Pair[string, string]{
	{Key: "report", Value: "Report"},
	{Key: "briefing", Value: "Briefing"},
	{Key: "memo", Value: "Memo"},
	{Key: "dataset", Value: "Dataset summary"},
	{Key: "other", Value: "Other"},
}

// Report is a manually curated document (reports, briefings, etc.) with an optional embedding for retrieval when [GeminiEmbeddingConfig] / [VLEmbedder] is configured.
type Report struct {
	gorm.Model
	Name        string `gorm:"not null"`
	Description string `gorm:"type:text"`
	// Type is a key from [ReportTypeChoices].
	Type    string `gorm:"not null"`
	Content string `gorm:"type:text;not null;default:''"`
	// Embedding matches [IntelEmbeddingDim] and the configured [VLEmbedder].
	Embedding *pgvector.Vector `gorm:"type:vector(1024)"`
}

func (Report) TableName() string { return "targets_of_interest" }

// String returns markdown-shaped text for display and for [VLEmbedder] input (name, type, description, content).
func (a *Report) String() string {
	if a == nil {
		return ""
	}
	var b strings.Builder
	if t := strings.TrimSpace(a.Name); t != "" {
		fmt.Fprintf(&b, "# %s\n\n", t)
	}
	if t := strings.TrimSpace(a.Type); t != "" {
		fmt.Fprintf(&b, "**Type:** %s\n\n", t)
	}
	if t := strings.TrimSpace(a.Description); t != "" {
		b.WriteString(t)
		b.WriteString("\n\n")
	}
	if t := strings.TrimSpace(a.Content); t != "" {
		b.WriteString(t)
	}
	return strings.TrimSpace(b.String())
}

// BeforeSave sets [Report.Embedding] from name/type/description/content ([prepareReportEmbeddingForSave]).
func (a *Report) BeforeSave(tx *gorm.DB) error {
	if tx.Statement.SkipHooks {
		return nil
	}
	return prepareReportEmbeddingForSave(context.Background(), a)
}

func init() {
	lago.OnDBInit("p_lacerate.report_model", func(db *gorm.DB) *gorm.DB {
		lago.RegisterModel[Report](db)
		return db
	})
}
