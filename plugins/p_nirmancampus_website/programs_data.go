package p_nirmancampus_website

import (
	"context"
	"log/slog"

	"github.com/lariv-in/lago/p_programs"
)

type programsPageData struct {
	Programs []websiteProgram
}

type websiteProgram struct {
	Name        string
	Code        string
	Description string
}

func buildProgramsPageData(ctx context.Context) programsPageData {
	db, err := homePageDB(ctx)
	if err != nil {
		slog.Error("nirmancampus_website: missing db while building programs page", "error", err)
		return programsPageData{}
	}

	var programs []p_programs.Program
	if err := db.Model(&p_programs.Program{}).
		Order("name ASC, code ASC").
		Find(&programs).Error; err != nil {
		slog.Error("nirmancampus_website: failed loading programs", "error", err)
		return programsPageData{}
	}

	items := make([]websiteProgram, 0, len(programs))
	for _, p := range programs {
		items = append(items, websiteProgram{
			Name:        p.Name,
			Code:        p.Code,
			Description: p.Description,
		})
	}

	return programsPageData{Programs: items}
}

