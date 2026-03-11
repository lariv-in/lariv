package lago

import "flag"

var (
	GenerateFlag = flag.Bool("generate", false, "Run data generators to seed the database")
	TuiFlag      = flag.Bool("run_tui", false, "Run the TUI")
)

func ParseFlags() {
	flag.Parse()
}
