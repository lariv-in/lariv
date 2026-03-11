package lago

import (
	"fmt"
	"log/slog"

	"gorm.io/gorm"
)

var RegistryGenerator Registry[Generator] = NewRegistry[Generator]()

type Generator struct {
	Create func(*gorm.DB)
	Remove func(*gorm.DB)
}

func RunGenerators() {
	db, err := InitDb()
	if err != nil {
		slog.Error("Failed to initialize database for generators", "error", err)
		return
	}

	generators := RegistryGenerator.All()
	for name, gen := range *generators {
		fmt.Printf("\n=== Running generator: %s ===\n", name)
		if gen.Remove != nil {
			gen.Remove(db)
		}
		if gen.Create != nil {
			gen.Create(db)
		}
	}
	fmt.Println("\nData generation complete.")
}
