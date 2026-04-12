package p_lacerate

import (
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/registry"
	"github.com/pgvector/pgvector-go"
	"gorm.io/gorm"
)

// ReportInterface is implemented per [Report.Kind].
type ReportInterface interface {
	reportKindRow()
	ReportString(*Report) string
	ReportSnippet(*Report) string
}

type ReportDesc struct {
	Name string
}

var ReportKindMap = map[string]ReportDesc{}

// ReportKindChoices is persisted key (Key) and UI label (Value) for [Report.Kind].
var ReportKindChoices = []registry.Pair[string, string]{
	{Key: "briefing", Value: "Briefing"},
	{Key: "timeline", Value: "Timeline"},
}

// RegistryReportKind holds a constructor per [Report.Kind] that returns a new kind row
// (for example `&BriefingReport{}`) for GORM to scan into.
var RegistryReportKind = registry.NewRegistry[func() ReportInterface]()

// Report is base row shared by all report kinds.
// Embedding text depends on kind-specific child rows, so keep it in sync from
// report-kind writes rather than relying on base-row fields alone.
type Report struct {
	gorm.Model
	Name        string `gorm:"not null"`
	Description string `gorm:"type:text"`
	Kind        string `gorm:"not null"`
	// Embedding is derived from base fields plus kind-specific child content.
	Embedding *pgvector.Vector `gorm:"type:vector(1024)"`
}

func (Report) TableName() string { return "reports" }

// deleteReportKindExtensionRows removes every per-kind row tied to this [Report].
// Call inside a transaction before deleting or replacing kind-specific data.
func deleteReportKindExtensionRows(tx *gorm.DB, reportID uint) error {
	if reportID == 0 {
		return nil
	}
	var timelines []TimelineReport
	if err := tx.Where("report_id = ?", reportID).Find(&timelines).Error; err != nil {
		return err
	}
	if len(timelines) != 0 {
		ids := make([]uint, 0, len(timelines))
		for _, row := range timelines {
			ids = append(ids, row.ID)
		}
		if err := tx.Where("timeline_report_id IN ?", ids).Delete(&TimelineReportEntry{}).Error; err != nil {
			return err
		}
	}
	if err := tx.Where("report_id = ?", reportID).Delete(&TimelineReport{}).Error; err != nil {
		return err
	}
	if err := tx.Where("report_id = ?", reportID).Delete(&BriefingReport{}).Error; err != nil {
		return err
	}
	return nil
}

func init() {
	lago.OnDBInit("p_lacerate.report_model", func(db *gorm.DB) *gorm.DB {
		lago.RegisterModel[Report](db)
		return db
	})
}
