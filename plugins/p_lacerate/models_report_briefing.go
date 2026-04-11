package p_lacerate

import (
	"context"
	"fmt"
	"strings"

	"github.com/lariv-in/lago/lago"
	"gorm.io/gorm"
)

// BriefingReport stores markdown body for a briefing-kind [Report].
type BriefingReport struct {
	gorm.Model
	ReportID uint   `gorm:"not null;uniqueIndex"`
	Report   Report `gorm:"foreignKey:ReportID"`
	Content  string `gorm:"type:text;not null;default:''"`
}

func (BriefingReport) TableName() string { return "report_briefings" }

func (*BriefingReport) reportKindRow() {}

func (b *BriefingReport) ReportString(report *Report) string {
	if report == nil {
		return ""
	}
	var sb strings.Builder
	if name := strings.TrimSpace(report.Name); name != "" {
		fmt.Fprintf(&sb, "# %s\n\n", name)
	}
	sb.WriteString("**Kind:** briefing\n\n")
	if desc := strings.TrimSpace(report.Description); desc != "" {
		sb.WriteString(desc)
		sb.WriteString("\n\n")
	}
	if b != nil {
		if body := strings.TrimSpace(b.Content); body != "" {
			sb.WriteString(body)
		}
	}
	return strings.TrimSpace(sb.String())
}

func (b *BriefingReport) ReportSnippet(report *Report) string {
	if report != nil {
		if desc := strings.TrimSpace(report.Description); desc != "" {
			return desc
		}
	}
	if b == nil {
		return ""
	}
	return strings.TrimSpace(b.Content)
}

func (b *BriefingReport) AfterSave(tx *gorm.DB) error {
	if tx.Statement.SkipHooks || b == nil || b.ReportID == 0 {
		return nil
	}
	return refreshReportEmbedding(context.Background(), tx, b.ReportID)
}

func init() {
	ReportKindMap["briefing"] = ReportDesc{Name: "Briefing"}
	if err := RegistryReportKind.Register("briefing", func() ReportInterface { return &BriefingReport{} }); err != nil {
		panic(err)
	}
	lago.OnDBInit("p_lacerate.report_briefing_model", func(db *gorm.DB) *gorm.DB {
		lago.RegisterModel[BriefingReport](db)
		return db
	})
}
