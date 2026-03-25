package main

import (
	"log/slog"

	"github.com/lariv-in/lago/lago"
	_ "github.com/lariv-in/lago/p_academicrecords"
	_ "github.com/lariv-in/lago/p_academicrecords_courses"
	_ "github.com/lariv-in/lago/p_academicrecords_programs"
	_ "github.com/lariv-in/lago/p_announcements"
	_ "github.com/lariv-in/lago/p_announcements_semesters"
	_ "github.com/lariv-in/lago/p_assignmentresults"
	_ "github.com/lariv-in/lago/p_assignments"
	_ "github.com/lariv-in/lago/p_assignments_semesters"
	_ "github.com/lariv-in/lago/p_courses"
	_ "github.com/lariv-in/lago/p_courses_teachers"
	_ "github.com/lariv-in/lago/p_dashboard"
	_ "github.com/lariv-in/lago/p_filesystem"
	_ "github.com/lariv-in/lago/p_nirmancampus_students"
	_ "github.com/lariv-in/lago/p_nirmancampus_users"
	_ "github.com/lariv-in/lago/p_nirmancampus_website"
	_ "github.com/lariv-in/lago/p_otp"
	_ "github.com/lariv-in/lago/p_programs"
	_ "github.com/lariv-in/lago/p_nirmancampus_programs"
	_ "github.com/lariv-in/lago/p_pwa"
	_ "github.com/lariv-in/lago/p_semesters"
	_ "github.com/lariv-in/lago/p_nirmancampus_studentapplications"
	_ "github.com/lariv-in/lago/p_students"
	_ "github.com/lariv-in/lago/p_teachers"
	_ "github.com/lariv-in/lago/p_users"
)

func main() {
	config, err := lago.LoadConfigFromFile("nirmancampus.toml")
	if err != nil {
		panic(err)
	}
	if err := lago.Start(config); err != nil {
		slog.Error(err.Error())
	}
}
