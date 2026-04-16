package main

import (
	"log/slog"

	"github.com/lariv-in/lago/lago"
	_ "github.com/lariv-in/lago/plugins/p_dashboard"
	_ "github.com/lariv-in/lago/plugins/p_filesystem"
	_ "github.com/lariv-in/lago/plugins/p_nirmancampus_academicrecords"
	_ "github.com/lariv-in/lago/plugins/p_nirmancampus_announcements"
	_ "github.com/lariv-in/lago/plugins/p_nirmancampus_assignmentsubmissions"
	_ "github.com/lariv-in/lago/plugins/p_nirmancampus_courses"
	_ "github.com/lariv-in/lago/plugins/p_nirmancampus_programs"
	_ "github.com/lariv-in/lago/plugins/p_nirmancampus_sessions"
	_ "github.com/lariv-in/lago/plugins/p_nirmancampus_students"
	_ "github.com/lariv-in/lago/plugins/p_nirmancampus_studentpayments"
	_ "github.com/lariv-in/lago/plugins/p_nirmancampus_users"
	_ "github.com/lariv-in/lago/plugins/p_nirmancampus_website"
	_ "github.com/lariv-in/lago/plugins/p_otp"
	_ "github.com/lariv-in/lago/plugins/p_pwa"
	_ "github.com/lariv-in/lago/plugins/p_users"
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
