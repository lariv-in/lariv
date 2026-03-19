package main

import (
	"log/slog"

	"github.com/lariv-in/lago"
	_ "github.com/lariv-in/p_dashboard"
	_ "github.com/lariv-in/p_filesystem"
	_ "github.com/lariv-in/p_otp"
	_ "github.com/lariv-in/p_pwa"
	_ "github.com/lariv-in/p_students"
	_ "github.com/lariv-in/p_teachers"
	_ "github.com/lariv-in/p_courses"
	_ "github.com/lariv-in/p_users"
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
