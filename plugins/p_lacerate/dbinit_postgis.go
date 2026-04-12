package p_lacerate

import (
	"log/slog"

	"github.com/lariv-in/lago/lago"
	"gorm.io/gorm"
)

func init() {
	lago.OnDBInit("p_lacerate.postgis_extension", func(db *gorm.DB) *gorm.DB {
		if db.Name() != "postgres" {
			return db
		}
		if err := db.Exec(`CREATE EXTENSION IF NOT EXISTS postgis`).Error; err != nil {
			slog.Error("lacerate: create postgis extension", "error", err)
		}
		return db
	})
}
