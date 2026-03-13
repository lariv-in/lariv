package lago

import "flag"

var (
	GenerateFlag = flag.Bool("generate", false, "Run data generators to seed the database")
)

func ParseFlags() {
	flag.Parse()
}
