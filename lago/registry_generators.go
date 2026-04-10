package lago

import (
	"fmt"
	"log/slog"

	"github.com/lariv-in/lago/registry"
	"gorm.io/gorm"
)

var RegistryGenerator *registry.Registry[Generator] = registry.NewRegistry[Generator]()

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

	if len(config.GeneratorOrder) == 0 {
		slog.Warn("InstalledApps is empty in config, running generators in non-deterministic order")
		for name, gen := range generators {
			fmt.Printf("\n=== Running generator: %s ===\n", name)
			if gen.Remove != nil {
				gen.Remove(db)
			}
			if gen.Create != nil {
				gen.Create(db)
			}
		}
		fmt.Println("\nData generation complete.")
		return
	}

	// Phase 1: Remove in reverse order (respects foreign key dependencies)
	fmt.Println("=== Phase 1: Removing generated data (reverse order) ===")
	for i := len(config.GeneratorOrder) - 1; i >= 0; i-- {
		name := config.GeneratorOrder[i]
		gen, ok := generators[name]
		if !ok {
			continue // No generator registered for this app
		}
		if gen.Remove != nil {
			fmt.Printf("  Removing: %s\n", name)
			if err := gen.Remove(db); err != nil {
				slog.Error("Generator remove failed", "name", name, "error", err)
			}
		}
	}

	// Phase 2: Create in forward order (dependencies created first)
	fmt.Println("\n=== Phase 2: Creating generated data (forward order) ===")
	for _, name := range config.GeneratorOrder {
		gen, ok := generators[name]
		if !ok {
			continue // No generator registered for this app
		}
		if gen.Create != nil {
			fmt.Printf("  Creating: %s\n", name)
			if err := gen.Create(db); err != nil {
				slog.Error("Generator create failed", "name", name, "error", err)
			}
		}
	}

	// Warn about generators not listed in InstalledApps
	listed := make(map[string]bool, len(config.GeneratorOrder))
	for _, name := range config.GeneratorOrder {
		listed[name] = true
	}
	for name := range generators {
		if !listed[name] {
			slog.Warn("Generator registered but not listed in InstalledApps, it will not run", "name", name)
		}
	}

	fmt.Println("\nData generation complete.")
}
