package lago

import (
	"log/slog"

	"github.com/lariv-in/lago/registry"
	"gorm.io/gorm"
)

var RegistryModel *registry.Registry[any] = registry.NewRegistry[any]()

func RegisterModel[T any](db *gorm.DB) {
	var zero T
	err := db.AutoMigrate(&zero)
	if err != nil {
		slog.Error("Error while migrating", "error", err)
	}
	stmt := &gorm.Statement{DB: db}
	err = stmt.Parse(&zero)
	if err != nil {
		slog.Error("Error while migrating", "error", err)
	}
	tableName := stmt.Schema.Table
	RegistryModel.Register(tableName, zero)
}
