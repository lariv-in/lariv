package main

import (
	"log/slog"

	"github.com/lariv-in/lago"
	_ "github.com/lariv-in/p_dashboard"
	_ "github.com/lariv-in/p_otp"
	_ "github.com/lariv-in/p_pwa"
	_ "github.com/lariv-in/p_totschool_appointments"
	_ "github.com/lariv-in/p_totschool_proposals"
	_ "github.com/lariv-in/p_totschool_tally"
	_ "github.com/lariv-in/p_totschool_users"
	_ "github.com/lariv-in/p_users"
)

func main() {
	config, err := lago.LoadConfigFromFile("totschool.toml")
	if err != nil {
		panic(err)
	}
	if err := lago.Start(config); err != nil {
		slog.Error(err.Error())
	}
}
