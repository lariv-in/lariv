package lago

import (
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var dbInitRegistry []func(*gorm.DB) *gorm.DB = []func(*gorm.DB) *gorm.DB{}

func OnDBInit(f func(*gorm.DB) *gorm.DB) {
	dbInitRegistry = append(dbInitRegistry, f)
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

	db, err := gorm.Open(dialector, &gorm.Config{})
	if err != nil {
		return nil, err
	}

	for _, f := range dbInitRegistry {
		db = f(db)
	}
	return db, nil
}
