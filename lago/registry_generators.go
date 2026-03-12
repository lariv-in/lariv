package lago

import (
	"fmt"
	"log/slog"

	"github.com/lariv-in/registry"
	"gorm.io/gorm"
)

var RegistryGenerator registry.Registry[Generator] = registry.NewRegistry[Generator]()

type Generator struct {
	Create func(*gorm.DB) error
	Remove func(*gorm.DB) error
}

func RunGenerators(config LagoConfig) {
	db, err := InitDB(config)
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
