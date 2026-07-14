package lago

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
	"github.com/lariv-in/lago/registry"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
)

// LagoConfig represents the top-level configuration structure mapped from TOML files.
// It carries connection details, database setups, UDS paths, CORS trusted origins, and plugin parameters.
type LagoConfig struct {
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

// LoadConfigFromFile decodes a TOML configuration file, registers application plugins,
// initializes database connections, decodes specific plugin configurations, and runs database migrations and hooks.
//
// Use Cases:
//   - Parsing configurations from files at startup before running the Cobra TUI or web server.
//
// Example:
//
//	config, err := lago.LoadConfigFromFile("config.toml", plugins)
//	if err != nil {
//		log.Fatal(err)
//	}
func LoadConfigFromFile(path string, plugins []registry.Pair[string, Plugin]) (LagoConfig, error) {
	var config LagoConfig

	if path == "" {
		return config, fmt.Errorf("config path is empty")
	}

	resolvedPath := path
	if !filepath.IsAbs(resolvedPath) {
		if _, err := os.Stat(resolvedPath); err == nil {
			// File exists in the current working directory, use it directly.
		} else {
			// Fallback to the directory of the binary.
			exe, err := os.Executable()
			if err != nil {
				slog.Error("failed resolving executable path for config file", "err", err, "configPath", path)
				return config, err
			}
			resolvedPath = filepath.Join(filepath.Dir(exe), resolvedPath)
		}
	}

	md, err := toml.DecodeFile(resolvedPath, &config)
	if err != nil {
		slog.Error("failed decoding config file", "err", err, "configPath", path, "resolvedPath", resolvedPath)
		return config, err
	}

	db, err := GetDbConn(config)
	if err != nil {
		return config, err
	}
	BuildAllRegistries(append([]registry.Pair[string, Plugin]{CorePlugin(db, config)}, plugins...))

	// Decode plugin configs before InitDB so DB-init hooks (e.g. background
	// workers started in a hook) observe fully populated config rather than
	// racing against config decoding.
	for key, cfgPointer := range RegistryConfig.All() {
		if prim, ok := config.Plugins[key]; ok {
			err = md.PrimitiveDecode(prim, cfgPointer)
			if err != nil {
				slog.Error("failed decoding plugin config", "err", err, "plugin", key)
				return config, err
			}
		}
		// Run even when the app has no [Plugins.<key>] table, so plugins can require fields
		// (e.g. panic if mandatory secrets are missing) instead of silently skipping validation.
		cfgPointer.PostConfig()
	}

	if err := InitDB(db, config); err != nil {
		return config, err
	}

	return config, nil
}
