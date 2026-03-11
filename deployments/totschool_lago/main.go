package main

import (
	"log/slog"

	"github.com/lariv-in/lago"
	_ "github.com/lariv-in/p_dashboard"
	_ "github.com/lariv-in/p_otp"
	_ "github.com/lariv-in/p_users"
)

func main() {
	lago.ParseFlags()

	if *lago.GenerateFlag {
		lago.RunGenerators()
		return
	}

	if *lago.TuiFlag {
		lago.RunTui()
	} else {
		slog.Error(lago.Start("127.0.0.1:42069", nil, nil).Error())
	}
}
