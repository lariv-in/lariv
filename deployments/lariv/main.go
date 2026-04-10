package main

import (
	"log/slog"

	"github.com/lariv-in/lago/lago"

	_ "github.com/lariv-in/lago/plugins/p_contacts"
	_ "github.com/lariv-in/lago/plugins/p_dashboard"
	_ "github.com/lariv-in/lago/plugins/p_export"
	_ "github.com/lariv-in/lago/plugins/p_filesystem"
	_ "github.com/lariv-in/lago/plugins/p_forms"
	_ "github.com/lariv-in/lago/plugins/p_lacerate"
	_ "github.com/lariv-in/lago/plugins/p_otp"
	_ "github.com/lariv-in/lago/plugins/p_pwa"
	_ "github.com/lariv-in/lago/plugins/p_users"
)

func main() {
	config, err := lago.LoadConfigFromFile("lariv.toml")
	if err != nil {
		panic(err)
	}
	if err := lago.Start(config); err != nil {
		slog.Error(err.Error())
	}
}
