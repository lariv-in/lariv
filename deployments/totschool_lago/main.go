package main

import (
	"log/slog"

	"github.com/lariv-in/lago"
	_ "github.com/lariv-in/p_courses"
	_ "github.com/lariv-in/p_dashboard"
	_ "github.com/lariv-in/p_otp"
	_ "github.com/lariv-in/p_totschool_users"
	_ "github.com/lariv-in/p_totschool_appointments"
	_ "github.com/lariv-in/p_totschool_proposals"
	_ "github.com/lariv-in/p_totschool_tally"
	_ "github.com/lariv-in/p_users"
)

func main() {
	lago.ParseFlags()

	config, err := lago.LoadConfigFromFile("totschool.toml")
	if err != nil {
		panic(err)
	}
	if *lago.GenerateFlag {
		lago.RunGenerators(config)
		return
	}

	if *lago.TuiFlag {
		lago.RunTui()
	} else {
		slog.Error(lago.Start(config).Error())
	}
}
