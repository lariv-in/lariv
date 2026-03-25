package p_nirmancampus_programs

import (
	"log"

	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_programs"
	"gorm.io/gorm"
)

// NirmancampusProgramDetails is a one-to-one extension of p_programs.Program.
type NirmancampusProgramDetails struct {
	gorm.Model

	ProgramID uint `gorm:"uniqueIndex;not null"`
	Program   p_programs.Program `gorm:"constraint:OnDelete:CASCADE;foreignKey:ProgramID;references:ID"`

	University string `gorm:"type:varchar(32);not null;default:''"`
}

func init() {
	lago.OnDBInit(func(d *gorm.DB) *gorm.DB {
		if err := d.AutoMigrate(&NirmancampusProgramDetails{}); err != nil {
			log.Panicf("failed to migrate NirmancampusProgramDetails: %v", err)
		}
		return d
	})
}
