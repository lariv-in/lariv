package p_nirmancampus_website

import (
	"context"
	"log/slog"

	"github.com/lariv-in/lago/plugins/p_nirmancampus_programs"
	"github.com/lariv-in/lago/plugins/p_programs"
)

type programsPageData struct {
	Programs []websiteProgram
}

type websiteProgram struct {
	Name        string
	Code        string
	Description string
	University  string
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

	universityByProgramID := map[uint]string{}
	if len(programs) > 0 {
		ids := make([]uint, len(programs))
		for i := range programs {
			ids[i] = programs[i].ID
		}
		var details []p_nirmancampus_programs.NirmancampusProgramDetails
		if err := db.Where("program_id IN ?", ids).Find(&details).Error; err != nil {
			slog.Error("nirmancampus_website: failed loading program university details", "error", err)
		} else {
			for _, d := range details {
				if d.University != "" {
					universityByProgramID[d.ProgramID] = d.University
				}
			}
		}
	}

	items := make([]websiteProgram, 0, len(programs))
	for _, p := range programs {
		items = append(items, websiteProgram{
			Name:        p.Name,
			Code:        p.Code,
			Description: p.Description,
			University:  universityByProgramID[p.ID],
		})
	}

	return programsPageData{Programs: items}
}

