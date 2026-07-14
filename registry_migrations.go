package lago

import (
	"github.com/lariv-in/lago/registry"
)

// RegistryMigrations represents the global immutable registry tracking database migrations filesystems of type [UsefulFilesystem].
// The database engine scans filesystems registered here to automatically execute pending SQL schemas during startup.
//
// Use Cases:
//   - Bundling plugin database table creations, constraints, or seeds.
//
// Example Registration:
//
//	//go:embed migrations/*.sql
//	var MigrationFS embed.FS
//
//	// In your lago.Plugin setup:
//	lago.Plugin{
//		Migrations: lago.PluginStages(func() PluginFeatures[UsefulFilesystem] {
//			return PluginFeatures[UsefulFilesystem]{
//				Entries: []registry.Pair[string, UsefulFilesystem]{
//					registry.NewPair("my_plugin_migrations", MigrationFS),
//				},
//			}
//		}),
//	}
//
// Example Patch:
//
//	// Register a patch to modify migration filesystems from another plugin:
//	lago.Plugin{
//		Migrations: lago.PluginStages(func() PluginFeatures[UsefulFilesystem] {
//			return PluginFeatures[UsefulFilesystem]{
//				Patches: []registry.Pair[string, func(UsefulFilesystem) UsefulFilesystem]{
//					registry.NewPair("my_plugin_migrations", func(existing UsefulFilesystem) UsefulFilesystem {
//						// Decorate or intercept filesystem:
//						return existing
//					}),
//				},
//			}
//		}),
//	}
//
// Example Retrieval:
//
//	fs, ok := RegistryMigrations.Get("my_plugin_migrations")
var RegistryMigrations *registry.ImmutableRegistry[UsefulFilesystem] = &registry.ImmutableRegistry[UsefulFilesystem]{}
