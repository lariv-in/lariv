package lago

import (
	"log"

	"github.com/lariv-in/lago/registry"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// DBInitHook runs after core DB setup (migrations, callbacks). Hooks run in registration order.
type DBInitHook func(*gorm.DB) *gorm.DB

// RegistryDBInit stores DB init hooks; iterate with [registry.RegisterOrder] to preserve registration order.
// [AllStable] returns an internal cached slice — do not mutate it.
var RegistryDBInit = registry.NewRegistry[DBInitHook]()

// OnDBInit registers hook under name. Duplicate names panic at init time.
func OnDBInit(name string, hook DBInitHook) {
	if err := RegistryDBInit.Register(name, hook); err != nil {
		log.Panic(err)
	}
}

func InitDB(config LagoConfig) (*gorm.DB, error) {
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
	})
	if err != nil {
		return nil, err
	}

	// Configure hard delete - skip soft delete and actually remove rows
	db.Callback().Delete().Before("gorm:delete").Register("lago:hard_delete", func(db *gorm.DB) {
		// Set Unscoped to true to force hard delete instead of soft delete
		db.Statement.Unscoped = true
	})

	for _, p := range *RegistryDBInit.AllStable(registry.RegisterOrder[DBInitHook]{}) {
		db = p.Value(db)
	}
	return db, nil
}
