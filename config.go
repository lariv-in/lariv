package lariv

import (
	"github.com/BurntSushi/toml"
	"github.com/lariv-in/lariv/registry"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
)

// LarivConfig represents the top-level configuration structure mapped from TOML files.
// It carries connection details, database setups, UDS paths, CORS trusted origins, and plugin parameters.
type LarivConfig struct {
	// Debug enables verbose debug level outputs and diagnostics.
	Debug bool
	// DBType specifies the database driver engine type (e.g. Postgres, Sqlite).
	DBType DBType
	// SqliteConfig represents driver parameters for SQLite DB files.
	SqliteConfig *sqlite.Config
	// PostgresConfig represents connection parameters for PostgreSQL connections.
	PostgresConfig *postgres.Config
	// Address represents the TCP bind address (e.g. ":8080").
	Address string
	// UDS represents the Unix Domain Socket path to bind to (overrides Address if specified).
	UDS string
	// GeneratorOrder specifies the sequence of db seeder names to run during seed execution.
	GeneratorOrder []string
	// TrustedOrigins lists the allowed CORS request origin hosts.
	TrustedOrigins []string
	// Plugins maps raw TOML configuration sections to specific plugin config structures.
	Plugins map[string]toml.Primitive
}

// DBType represents the configuration database engine driver selector.
type DBType string

const (
	// DBTypeSqlite specifies GORM SQLite database configurations.
	DBTypeSqlite = DBType("Sqlite")
	// DBTypePostgres specifies GORM PostgreSQL database configurations.
	DBTypePostgres = DBType("Postgres")
)

// Config defines the interface implemented by plugin config structs to receive and validate parsed settings from TOML files.
// PostConfig is executed automatically after settings are mapped, enabling validation or setting default values.
//
// Use Cases:
//   - Defining configuration tables for plugins (e.g. storage paths, API client secrets).
//
// Example Definition:
//
//	type DashboardConfig struct {
//		AppName string
//	}
//
//	func (c *DashboardConfig) PostConfig() {
//		if c.AppName == "" {
//			c.AppName = "My Dashboard App"
//		}
//	}
//
// Example Registration:
//
//	var DashboardConfigPtr = &DashboardConfig{}
//
//	// Register the config instance inside your lariv.Plugin configuration:
//	lariv.Plugin{
//		Configs: lariv.PluginStages(func() PluginFeatures[Config] {
//			return PluginFeatures[Config]{
//				Entries: []registry.Pair[string, Config]{
//					registry.NewPair("dashboard", DashboardConfigPtr),
//				},
//			}
//		}),
//	}
type Config interface {
	// PostConfig executes sanity checks and assigns default values after TOML values are loaded.
	PostConfig()
}

// LoadConfigFromFile decodes a TOML configuration file via [AppBuilder] and returns the config.
// Prefer NewBuilder().AddPlugins(plugins).LoadConfigFromFile(path) to retain the [*App].
func LoadConfigFromFile(path string, plugins []registry.Pair[string, Plugin]) (LarivConfig, error) {
	app, err := NewBuilder().AddPlugins(plugins).LoadConfigFromFile(path)
	if err != nil {
		return LarivConfig{}, err
	}
	return app.Config, nil
}
