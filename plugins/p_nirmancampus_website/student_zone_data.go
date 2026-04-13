package p_nirmancampus_website

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/lariv-in/lago/getters"
	"gorm.io/gorm"
)

type studentZonePageData struct {
	Sections []studentZoneSection
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
	db, err := getters.DBFromContext(ctx)
	if err != nil {
		slog.Error("nirmancampus_website: missing db while building student zone page", "error", err)
		return studentZonePageData{}
	}

	sections, err := gorm.G[StudentZoneSection](db).Order(`"order" ASC`).Find(ctx)
	if err != nil {
		slog.Error("nirmancampus_website: failed loading student zone sections", "error", err)
		return studentZonePageData{}
	}

	result := make([]studentZoneSection, 0, len(sections))
	for _, s := range sections {
		items, err := gorm.G[StudentZoneItem](db).Where("student_zone_section_id = ?", s.ID).Find(ctx)
		if err != nil {
			slog.Error("nirmancampus_website: failed loading student zone items", "error", err, "section_id", s.ID)
			continue
		}

		sectionItems := make([]studentZoneItem, 0, len(items))
		for _, item := range items {
			sectionItems = append(sectionItems, studentZoneItem{
				Title: item.Title,
				URL:   fmt.Sprintf("%s%d/", StudentZoneItemURLPrefix, item.ID),
			})
		}

		result = append(result, studentZoneSection{
			Title: s.Title,
			Items: sectionItems,
		})
	}

	return studentZonePageData{
		Sections: result,
	}
}
