package lariv

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/lariv-in/lariv/registry"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DBInitHook represents a hook function executed after core database setup is complete (migrations, default callbacks).
// Hooks are executed in registration order and receive the active *gorm.DB instance, returning a decorated/modified *gorm.DB client.
//
// Use Cases:
//   - Configuring connection pool properties (e.g., setting maximum connections).
//   - Triggering initial auto-migrations for non-plugin core tables.
//   - Running database logging configurations or starting background cron workers with database handles.
//
// Example Definition:
//
//	var ConnPoolHook DBInitHook = func(db *gorm.DB) *gorm.DB {
//		sqlDB, err := db.DB()
//		if err == nil {
//			sqlDB.SetMaxOpenConns(20)
//		}
//		return db
//	}
//
// Example Registration:
//
//	// In your lariv.Plugin setup:
//	lariv.Plugin{
//		DBInitHooks: lariv.PluginStages(func() PluginFeatures[DBInitHook] {
//			return PluginFeatures[DBInitHook]{
//				Entries: []registry.Pair[string, DBInitHook]{
//					registry.NewPair("conn_pool", ConnPoolHook),
//				},
//			}
//		}),
//	}
//
// Example Patch:
//
//	// Register a patch to chain or decorate existing DBInitHooks from another plugin:
//	lariv.Plugin{
//		DBInitHooks: lariv.PluginStages(func() PluginFeatures[DBInitHook] {
//			return PluginFeatures[DBInitHook]{
//				Patches: []registry.Pair[string, func(DBInitHook) DBInitHook]{
//					registry.NewPair("conn_pool", func(existing DBInitHook) DBInitHook {
//						return func(db *gorm.DB) *gorm.DB {
//							db = existing(db)
//							// Chain extra configurations:
//							return db.Debug()
//						}
//					}),
//				},
//			}
//		}),
//	}
//
// Example Retrieval:
//
//	hook, ok := RegistryDBInit.Get("conn_pool")
type DBInitHook func(*gorm.DB) *gorm.DB

// RegistryDBInit represents the global immutable registry tracking database initialization hooks.
var RegistryDBInit *registry.ImmutableRegistry[DBInitHook] = &registry.ImmutableRegistry[DBInitHook]{}

// GetDbConn opens and configures a GORM connection using the provided database configurations.
// It overrides default delete callbacks to enforce hard deletes (disabling soft deletes).
func GetDbConn(config LarivConfig) (*gorm.DB, error) {
	var dialector gorm.Dialector

	switch config.DBType {
	case DBTypeSqlite:
		dialector = sqlite.New(*config.SqliteConfig)
	case DBTypePostgres:
		dialector = postgres.New(*config.PostgresConfig)
	default:
		log.Panicf("Unrecognized db type %s", config.DBType)
	}

	db, err := gorm.Open(dialector, &gorm.Config{
		PrepareStmt: true,
		Logger: logger.New(
			log.Default(),
			logger.Config{
				SlowThreshold:             time.Hour * 24, // Practically disables slow log
				LogLevel:                  logger.Warn,
				IgnoreRecordNotFoundError: true,
				Colorful:                  true,
			},
		),
	})
	if err != nil {
		return nil, err
	}

	// Configure hard delete - skip soft delete and actually remove rows
	db.Callback().Delete().Before("gorm:delete").Register("lariv:hard_delete", func(db *gorm.DB) {
		// Set Unscoped to true to force hard delete instead of soft delete
		db.Statement.Unscoped = true
	})
	return db, nil
}

// InitDB executes pending schema migrations and invokes registered database initialization hooks sequentially.
func InitDB(db *gorm.DB, config LarivConfig) error {
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("gorm.DB().DB(): %w", err)
	}

	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetMaxIdleConns(25)
	sqlDB.SetConnMaxLifetime(1 * time.Hour)
	sqlDB.SetConnMaxIdleTime(15 * time.Minute)

	if err := gooseUpPluginMigrations(context.Background(), sqlDB, config); err != nil {
		return fmt.Errorf("goose migrations: %w", err)
	}

	for _, p := range *RegistryDBInit.AllStable() {
		db = p.Value(db)
	}
	return nil
}
