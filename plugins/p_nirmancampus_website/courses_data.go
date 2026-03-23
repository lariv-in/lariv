package p_nirmancampus_website

import (
	"context"
	"log/slog"

	"github.com/lariv-in/lago/p_courses"
)

type coursesPageData struct {
	Courses []websiteCourse
}

type websiteCourse struct {
	Name        string
	Code        string
	Level       string
	Subject     string
	Description string
}

func buildCoursesPageData(ctx context.Context) coursesPageData {
	db, err := homePageDB(ctx)
	if err != nil {
		slog.Error("nirmancampus_website: missing db while building courses page", "error", err)
		return coursesPageData{}
	}

	var courses []p_courses.Course
	if err := db.Model(&p_courses.Course{}).
		Where("is_active = ?", true).
		Order("level ASC, name ASC").
		Find(&courses).Error; err != nil {
		slog.Error("nirmancampus_website: failed loading courses", "error", err)
		return coursesPageData{}
	}

	items := make([]websiteCourse, 0, len(courses))
	for _, c := range courses {
		items = append(items, websiteCourse{
			Name:        c.Name,
			Code:        c.Code,
			Level:       c.Level,
			Subject:     c.Subject,
			Description: c.Description,
		})
	}

	return coursesPageData{Courses: items}
}
