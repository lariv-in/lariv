package main

import (
	"log/slog"

	"github.com/lariv-in/lago/lago"

	_ "github.com/lariv-in/lago/plugins/p_admissions"
	_ "github.com/lariv-in/lago/plugins/p_allocation"
	_ "github.com/lariv-in/lago/plugins/p_announcements"
	_ "github.com/lariv-in/lago/plugins/p_assessments"
	_ "github.com/lariv-in/lago/plugins/p_assignments"
	_ "github.com/lariv-in/lago/plugins/p_attendance"
	_ "github.com/lariv-in/lago/plugins/p_courses"
	_ "github.com/lariv-in/lago/plugins/p_dashboard"
	_ "github.com/lariv-in/lago/plugins/p_events"
	_ "github.com/lariv-in/lago/plugins/p_export"
	_ "github.com/lariv-in/lago/plugins/p_filesystem"
	_ "github.com/lariv-in/lago/plugins/p_finances"
	_ "github.com/lariv-in/lago/plugins/p_forms"
	_ "github.com/lariv-in/lago/plugins/p_forums"
	_ "github.com/lariv-in/lago/plugins/p_livereloading"
	_ "github.com/lariv-in/lago/plugins/p_otp"
	_ "github.com/lariv-in/lago/plugins/p_programs"
	_ "github.com/lariv-in/lago/plugins/p_pwa"
	_ "github.com/lariv-in/lago/plugins/p_reports"
	_ "github.com/lariv-in/lago/plugins/p_sarathi_institute"
	_ "github.com/lariv-in/lago/plugins/p_semesters"
	_ "github.com/lariv-in/lago/plugins/p_sessions"
	_ "github.com/lariv-in/lago/plugins/p_students"
	_ "github.com/lariv-in/lago/plugins/p_syllabus"
	_ "github.com/lariv-in/lago/plugins/p_teachers"
	_ "github.com/lariv-in/lago/plugins/p_timetable"
	_ "github.com/lariv-in/lago/plugins/p_users"
)

func main() {
	config, err := lago.LoadConfigFromFile("sarathi.toml")
	if err != nil {
		panic(err)
	}
	if err := lago.Start(config); err != nil {
		slog.Error(err.Error())
	}
}
