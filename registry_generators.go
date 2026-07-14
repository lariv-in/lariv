package lago

import (
	"fmt"
	"log/slog"

	"github.com/lariv-in/lago/registry"
	"gorm.io/gorm"
)

// RegistryGenerator represents the global immutable registry tracking database seed generators.
var RegistryGenerator *registry.ImmutableRegistry[Generator] = &registry.ImmutableRegistry[Generator]{}

// Generator defines creation and deletion functions executed during database seeding/data generating commands.
//
// Use Cases:
//   - Seeding development or test databases with mock tables data (e.g. creating dummy administrator accounts, catalog samples).
//
// Example Definition:
//
//	var ProductGen = Generator{
//		Create: func(db *gorm.DB) error {
//			return db.Create(&Product{Name: "Mock Item"}).Error
//		},
//		Remove: func(db *gorm.DB) error {
//			return db.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&Product{}).Error
//		},
//	}
//
// Example Registration:
//
//	// In your lago.Plugin setup:
//	lago.Plugin{
//		Generators: lago.PluginStages(func() PluginFeatures[Generator] {
//			return PluginFeatures[Generator]{
//				Entries: []registry.Pair[string, Generator]{
//					registry.NewPair("products_seeder", ProductGen),
//				},
//			}
//		}),
//	}
//
// Example Patch:
//
//	// Register a patch to chain or modify database seeders from another plugin:
//	lago.Plugin{
//		Generators: lago.PluginStages(func() PluginFeatures[Generator] {
//			return PluginFeatures[Generator]{
//				Patches: []registry.Pair[string, func(Generator) Generator]{
//					registry.NewPair("products_seeder", func(existing Generator) Generator {
//						return Generator{
//							Create: func(db *gorm.DB) error {
//								if err := existing.Create(db); err != nil {
//									return err
//								}
//								// Add extra seeder child data:
//								return db.Create(&Tag{Name: "Hot"}).Error
//							},
//							Remove: existing.Remove,
//						}
//					}),
//				},
//			}
//		}),
//	}
//
// Example Retrieval:
//
//	gen, ok := RegistryGenerator.Get("products_seeder")
type Generator struct {
	// Create populates database tables with mock data records.
	Create func(*gorm.DB) error
	// Remove cleans up database tables to undo mock population.
	Remove func(*gorm.DB) error
}

// RunGenerators executes all registered seed generators.
// It runs deletion (Remove) in reverse order of GeneratorOrder to satisfy foreign key constraints,
// and creation (Create) in forward order of GeneratorOrder so dependent entities are populated correctly.
func RunGenerators(config LagoConfig) {
	db, err := GetDbConn(config)
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
