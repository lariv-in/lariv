package main

import (
	"flag"
	"log/slog"

	"github.com/lariv-in/lago"
	_ "github.com/lariv-in/p_dashboard"
	_ "github.com/lariv-in/p_otp"
	_ "github.com/lariv-in/p_users"
)

func main() {
	runTui := flag.Bool("run_tui", false, "Run the tui")
	if *runTui {
		lago.RunTui()
	} else {
		slog.Error(lago.Start("127.0.0.1:42069", nil, nil).Error())
	}
}
