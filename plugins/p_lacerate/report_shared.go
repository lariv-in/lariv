package p_lacerate

import (
	"context"
	"fmt"
	"strings"

	"github.com/lariv-in/lago/registry"
	"gorm.io/gorm"
)

type ReportPageData struct {
	Report   Report
	Briefing *BriefingReport
	Timeline *TimelineReport
}

func reportKindLabel(kind string) string {
	if p, ok := registry.PairFromPairs(kind, ReportKindChoices); ok {
		return p.Value
	}
	return kind
}

func reportPageDataString(data ReportPageData) string {
	switch data.Report.Kind {
	case "briefing":
		if data.Briefing == nil {
			return ""
		}
		return data.Briefing.ReportString(&data.Report)
	case "timeline":
		if data.Timeline == nil {
			return ""
		}
		return data.Timeline.ReportString(&data.Report)
	default:
		return ""
	}
}

func reportPageDataSnippet(data ReportPageData) string {
	var s string
	switch data.Report.Kind {
	case "briefing":
		if data.Briefing != nil {
			s = data.Briefing.ReportSnippet(&data.Report)
		}
	case "timeline":
		if data.Timeline != nil {
			s = data.Timeline.ReportSnippet(&data.Report)
		}
	}
	s = strings.TrimSpace(s)
	if len(s) > 240 {
		return s[:237] + "..."
	}
	return s
}

func loadReportPageData(ctx context.Context, db *gorm.DB, reportID uint) (ReportPageData, error) {
	var data ReportPageData
	if err := db.WithContext(ctx).First(&data.Report, reportID).Error; err != nil {
		return data, err
	}
	newRow, ok := RegistryReportKind.Get(data.Report.Kind)
	if !ok {
		return data, fmt.Errorf("unsupported report kind %q", data.Report.Kind)
	}
	row := newRow()
	switch typed := row.(type) {
	case *BriefingReport:
		if err := db.WithContext(ctx).Where("report_id = ?", reportID).First(typed).Error; err != nil {
			return data, err
		}
		typed.Report = data.Report
		data.Briefing = typed
	case *TimelineReport:
		if err := db.WithContext(ctx).
			Preload("Entries", func(tx *gorm.DB) *gorm.DB {
				return tx.Order("datetime ASC").Order("position ASC").Order("id ASC")
			}).
			Where("report_id = ?", reportID).
			First(typed).Error; err != nil {
			return data, err
		}
		typed.Report = data.Report
		data.Timeline = typed
	default:
		return data, fmt.Errorf("unsupported report kind row %T", row)
	}
	return data, nil
}

func loadReportPageDataList(ctx context.Context, db *gorm.DB, reports []Report) ([]ReportPageData, error) {
	if len(reports) == 0 {
		return nil, nil
	}
	reportIDs := make([]uint, 0, len(reports))
	for _, report := range reports {
		reportIDs = append(reportIDs, report.ID)
	}

	var briefingRows []BriefingReport
	if err := db.WithContext(ctx).Where("report_id IN ?", reportIDs).Find(&briefingRows).Error; err != nil {
		return nil, err
	}
	briefingByReportID := make(map[uint]BriefingReport, len(briefingRows))
	for _, row := range briefingRows {
		briefingByReportID[row.ReportID] = row
	}

	var timelineRows []TimelineReport
	if err := db.WithContext(ctx).Where("report_id IN ?", reportIDs).Find(&timelineRows).Error; err != nil {
		return nil, err
	}
	timelineByReportID := make(map[uint]TimelineReport, len(timelineRows))
	timelineIDs := make([]uint, 0, len(timelineRows))
	for _, row := range timelineRows {
		timelineByReportID[row.ReportID] = row
		timelineIDs = append(timelineIDs, row.ID)
	}
	entriesByTimelineID := map[uint][]TimelineReportEntry{}
	if len(timelineIDs) != 0 {
		var entries []TimelineReportEntry
		if err := db.WithContext(ctx).
			Where("timeline_report_id IN ?", timelineIDs).
			Order("datetime ASC").
			Order("position ASC").
			Order("id ASC").
			Find(&entries).Error; err != nil {
			return nil, err
		}
		for _, entry := range entries {
			entriesByTimelineID[entry.TimelineReportID] = append(entriesByTimelineID[entry.TimelineReportID], entry)
		}
	}

	items := make([]ReportPageData, 0, len(reports))
	for _, report := range reports {
		item := ReportPageData{Report: report}
		if row, ok := briefingByReportID[report.ID]; ok {
			row.Report = report
			item.Briefing = &row
		}
		if row, ok := timelineByReportID[report.ID]; ok {
			row.Report = report
			row.Entries = entriesByTimelineID[row.ID]
			item.Timeline = &row
		}
		items = append(items, item)
	}
	return items, nil
}
