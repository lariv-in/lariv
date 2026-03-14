package lago

import (
	"flag"

	tea "charm.land/bubbletea/v2"
)

var tuiFlag = flag.Bool("tui", false, "Launch the tui instead of running the server")

func Start(config LagoConfig) error {
	if *tuiFlag {
		db, err := InitDB(config)
		if err != nil {
			return err
		}
		_, err = tea.NewProgram(initialModel(db)).Run()
		return err
	} else {
		return StartServer(config)
	}
}
