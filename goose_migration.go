package lago

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io/fs"
	"strings"
	"unicode"

	"github.com/pressly/goose/v3"
)

// GooseVersionTableName derives a database-safe schema version tracker table name for a plugin migration registry key.
// It maps plugin keys deterministically to clean snake_case table names (prefixed with "goose_migrations__")
// ensuring separate tables per plugin so that numeric migration versions can overlap across plugins.
//
// # Database Migrations Structure
//
// In Lago, database schema migrations are managed per plugin. To add migrations to a plugin:
//
// 1. Create a `migrations` folder inside your plugin directory structure.
// 2. Add SQL migration files inside the `migrations` directory following Goose's syntax structure:
//
//	-- 00001_create_users_table.sql
//	-- +goose Up
//	CREATE TABLE users (
//	    id SERIAL PRIMARY KEY,
//	    email VARCHAR(255) NOT NULL UNIQUE
//	);
//
//	-- +goose Down
//	DROP TABLE users;
//
// 3. Embed the migration files and register the filesystem on your [Plugin] configuration under the Migrations field:
//
//	//go:embed migrations/*.sql
//	var migrationFS embed.FS
//
//	// In your lago.Plugin setup:
//	lago.Plugin{
//	    Migrations: lago.PluginStages(func() lago.PluginFeatures[lago.UsefulFilesystem] {
//	        return lago.PluginFeatures[lago.UsefulFilesystem]{
//	            Entries: []registry.Pair[string, lago.UsefulFilesystem]{
//	                registry.NewPair("my_plugin", migrationFS),
//	            },
//	        }
//	    }),
//	}
func GooseVersionTableName(registryKey string) string {
	var b strings.Builder
	b.WriteString("goose_migrations__")
	prevUnderscore := false
	for _, r := range registryKey {
		switch {
		case unicode.IsLetter(r) || unicode.IsDigit(r):
			b.WriteRune(unicode.ToLower(r))
			prevUnderscore = false
		case r == '.' || r == '-' || r == '_':
			if !prevUnderscore {
				b.WriteByte('_')
				prevUnderscore = true
			}
		default:
			if !prevUnderscore {
				b.WriteByte('_')
				prevUnderscore = true
			}
		}
	}
	s := strings.Trim(b.String(), "_")
	if len(s) > len("goose_migrations__") {
		return s
	}
	return "goose_migrations__default"
}

func gooseDialect(t DBType) (goose.Dialect, error) {
	switch t {
	case DBTypeSqlite:
		return goose.DialectSQLite3, nil
	case DBTypePostgres:
		return goose.DialectPostgres, nil
	default:
		return "", fmt.Errorf("unsupported DBType for goose: %v", t)
	}
}

// gooseUpPluginMigrations cycles through all registered plugin migration folders, loading their target versions
// from their specific goose table names, and runs the "Up" migrations to update the database schema.
func gooseUpPluginMigrations(ctx context.Context, sqlDB *sql.DB, config LagoConfig) error {
	pairs := *RegistryMigrations.AllStable()
	if len(pairs) == 0 {
		return nil
	}
	dialect, err := gooseDialect(config.DBType)
	if err != nil {
		return err
	}
	for _, pair := range pairs {
		sub, err := fs.Sub(pair.Value, "migrations")
		if err != nil {
			return fmt.Errorf("migrations subdirectory for %q: %w", pair.Key, err)
		}
		p, err := goose.NewProvider(
			dialect,
			sqlDB,
			sub,
			goose.WithTableName(GooseVersionTableName(pair.Key)),
			goose.WithDisableGlobalRegistry(true),
		)
		if errors.Is(err, goose.ErrNoMigrations) {
			continue
		}
		if err != nil {
			return fmt.Errorf("goose NewProvider(%q): %w", pair.Key, err)
		}
		if _, err := p.Up(ctx); err != nil {
			return fmt.Errorf("goose Up for %q: %w", pair.Key, err)
		}
	}
	return nil
}
