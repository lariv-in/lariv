package p_nirmancampus_website

import (
	"context"
	"log/slog"
	"strings"

	"github.com/lariv-in/lago/p_courses"
)

type coursesPageData struct {
	Courses       []websiteCourse
	Levels        []string
	SelectedLevel string
}

type websiteCourse struct {
	Name        string
	Code        string
	Level       string
	Subject     string
	Description string
}

func coursesLevelFromContext(ctx context.Context) string {
	getMap, ok := ctx.Value("$get").(map[string]any)
	if !ok {
		return ""
	}
	raw, ok := getMap["Level"]
	if !ok || raw == nil {
		return ""
	}
	s, ok := raw.(string)
	if !ok {
		return ""
	}
	return strings.TrimSpace(s)
}

func buildCoursesPageData(ctx context.Context) coursesPageData {
	db, err := homePageDB(ctx)
	if err != nil {
		slog.Error("nirmancampus_website: missing db while building courses page", "error", err)
		return coursesPageData{}
	}

	selected := coursesLevelFromContext(ctx)

	var levels []string
	if err := db.Model(&p_courses.Course{}).
		Where("is_active = ?", true).
		Where("level IS NOT NULL AND level != ?", "").
		Distinct("level").
		Order("level ASC").
		Pluck("level", &levels).Error; err != nil {
		slog.Error("nirmancampus_website: failed loading course levels", "error", err)
		levels = nil
	}

	q := db.Model(&p_courses.Course{}).
		Where("is_active = ?", true)
	if selected != "" {
		q = q.Where("level = ?", selected)
	}

	var courses []p_courses.Course
	if err := q.Order("level ASC").Order("name ASC").Find(&courses).Error; err != nil {
		slog.Error("nirmancampus_website: failed loading courses", "error", err)
		return coursesPageData{Levels: levels, SelectedLevel: selected}
	}

	items := make([]websiteCourse, 0, len(courses))
	for _, course := range courses {
		items = append(items, websiteCourse{
			Name:        course.Name,
			Code:        course.Code,
			Level:       course.Level,
			Subject:     course.Subject,
			Description: course.Description,
		})
	}

	return coursesPageData{
		Courses:       items,
		Levels:        levels,
		SelectedLevel: selected,
	}
}
