package p_lacerate

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/lariv-in/lago/lago"
	"gorm.io/gorm"
)

// TimelineReport marks a timeline-kind [Report]; entries live in [TimelineReportEntry].
type TimelineReport struct {
	gorm.Model
	ReportID uint                  `gorm:"not null;uniqueIndex"`
	Report   Report                `gorm:"foreignKey:ReportID"`
	Entries  []TimelineReportEntry `gorm:"foreignKey:TimelineReportID"`
}

func (TimelineReport) TableName() string { return "report_timelines" }

func (*TimelineReport) reportKindRow() {}

func (t *TimelineReport) ReportString(report *Report) string {
	if report == nil {
		return ""
	}
	var sb strings.Builder
	if name := strings.TrimSpace(report.Name); name != "" {
		fmt.Fprintf(&sb, "# %s\n\n", name)
	}
	sb.WriteString("**Kind:** timeline\n\n")
	if desc := strings.TrimSpace(report.Description); desc != "" {
		sb.WriteString(desc)
		sb.WriteString("\n\n")
	}
	if t != nil && len(t.Entries) != 0 {
		sb.WriteString("## Entries\n\n")
		for i, entry := range t.Entries {
			if i != 0 {
				sb.WriteString("\n\n")
			}
			fmt.Fprintf(&sb, "### %s — %s\n\n", entry.Datetime.UTC().Format(time.RFC3339), strings.TrimSpace(entry.Title))
			if body := strings.TrimSpace(entry.Content); body != "" {
				sb.WriteString(body)
			}
		}
	}
	return strings.TrimSpace(sb.String())
}

func (t *TimelineReport) ReportSnippet(report *Report) string {
	if t != nil && len(t.Entries) != 0 {
		first := t.Entries[0]
		return strings.TrimSpace(fmt.Sprintf("%s — %s", first.Datetime.UTC().Format(time.RFC3339), first.Title))
	}
	if report == nil {
		return ""
	}
	return strings.TrimSpace(report.Description)
}

// TimelineReportEntry is one ordered timeline entry for a [TimelineReport].
type TimelineReportEntry struct {
	gorm.Model
	TimelineReportID uint           `gorm:"not null;index"`
	TimelineReport   TimelineReport `gorm:"foreignKey:TimelineReportID"`
	Position         uint           `gorm:"not null;default:0"`
	Datetime         time.Time      `gorm:"not null"`
	Title            string         `gorm:"not null;default:''"`
	Content          string         `gorm:"type:text;not null;default:''"`
}

func (TimelineReportEntry) TableName() string { return "report_timeline_entries" }

func (t *TimelineReport) AfterSave(tx *gorm.DB) error {
	if tx.Statement.SkipHooks || t == nil || t.ReportID == 0 {
		return nil
	}
	return refreshReportEmbedding(context.Background(), tx, t.ReportID)
}

func (e *TimelineReportEntry) AfterSave(tx *gorm.DB) error {
	if tx.Statement.SkipHooks || e == nil || e.TimelineReportID == 0 {
		return nil
	}
	var reportID uint
	if err := tx.Model(&TimelineReport{}).
		Where("id = ?", e.TimelineReportID).
		Select("report_id").
		Scan(&reportID).Error; err != nil {
		return err
	}
	if reportID == 0 {
		return nil
	}
	return refreshReportEmbedding(context.Background(), tx, reportID)
}

func init() {
	ReportKindMap["timeline"] = ReportDesc{Name: "Timeline"}
	if err := RegistryReportKind.Register("timeline", func() ReportInterface { return &TimelineReport{} }); err != nil {
		panic(err)
	}
	lago.OnDBInit("p_lacerate.report_timeline_model", func(db *gorm.DB) *gorm.DB {
		lago.RegisterModel[TimelineReport](db)
		lago.RegisterModel[TimelineReportEntry](db)
		return db
	})
}
