package p_nirmancampus_website

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	p_nirmancampus_student_zone "github.com/lariv-in/lago/p_nirmancampus_student_zone"
	"gorm.io/gorm"
)

type studentZonePageData struct {
	Announcements []homeAnnouncement
	Sections      []studentZoneSection
}

type studentZoneSection struct {
	Title string
	Items []studentZoneItem
}

type studentZoneItem struct {
	Title string
	URL   string
}

func buildStudentZonePageData(ctx context.Context) studentZonePageData {
	db, ok := ctx.Value("$db").(*gorm.DB)
	if !ok || db == nil {
		slog.Error("nirmancampus_website: missing db while building student zone page")
		return studentZonePageData{}
	}

	var sections []p_nirmancampus_student_zone.StudentZoneSection
	if err := db.Order(`"order" ASC`).Find(&sections).Error; err != nil {
		slog.Error("nirmancampus_website: failed loading student zone sections", "error", err)
		return studentZonePageData{}
	}

	result := make([]studentZoneSection, 0, len(sections))
	for _, s := range sections {
		var items []p_nirmancampus_student_zone.StudentZoneItem
		if err := db.Where("student_zone_section_id = ?", s.ID).Find(&items).Error; err != nil {
			slog.Error("nirmancampus_website: failed loading student zone items", "error", err, "section_id", s.ID)
			continue
		}

		sectionItems := make([]studentZoneItem, 0, len(items))
		for _, item := range items {
			sectionItems = append(sectionItems, studentZoneItem{
				Title: item.Title,
				URL:   fmt.Sprintf("/students-zone/item/%d/", item.ID),
			})
		}

		result = append(result, studentZoneSection{
			Title: s.Title,
			Items: sectionItems,
		})
	}

	return studentZonePageData{
		Announcements: loadHomeAnnouncements(db, time.Now()),
		Sections:      result,
	}
}
