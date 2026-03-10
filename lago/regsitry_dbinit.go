package lago

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var dbInitRegistry []func(*gorm.DB) *gorm.DB = []func(*gorm.DB) *gorm.DB{}

func OnDbInit(f func(*gorm.DB) *gorm.DB) {
	dbInitRegistry = append(dbInitRegistry, f)
}

func InitDb() (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	for _, f := range dbInitRegistry {
		db = f(db)
	}
	return db, nil
}
