package p_seer_intel

import (
	"log/slog"

	"github.com/lariv-in/lago/lago"
	"gorm.io/gorm"
)

func init() {
	lago.OnDBInit("p_seer_intel.pgvector_extension", func(db *gorm.DB) *gorm.DB {
		if db.Name() != "postgres" {
			return db
		}
		if err := db.Exec(`CREATE EXTENSION IF NOT EXISTS vector`).Error; err != nil {
			slog.Error("p_seer_intel: create pgvector extension", "error", err)
		}
		return db
	})
}
