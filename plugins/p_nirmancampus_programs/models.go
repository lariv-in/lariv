package p_nirmancampus_programs

import (
	"log"

	"github.com/lariv-in/lago/lago"
	"gorm.io/gorm"
)

// Program is the single programs row for Nirmancampus (replaces p_programs.Program + NirmancampusProgramDetails).
//
// If you already had data in nirmancampus_program_details, after AutoMigrate adds programs.university (PostgreSQL):
//
//	UPDATE programs AS p
//	SET university = d.university
//	FROM nirmancampus_program_details AS d
//	WHERE d.program_id = p.id AND d.deleted_at IS NULL;
//	DROP TABLE nirmancampus_program_details;
//
// Confirm table/column names in your database before running.
type Program struct {
	gorm.Model

	Name        string
	Code        string `gorm:"uniqueIndex"`
	Description string
	University  string `gorm:"type:varchar(32);not null;default:''"`
}

func init() {
	lago.OnDBInit(func(d *gorm.DB) *gorm.DB {
		if err := d.AutoMigrate(&Program{}); err != nil {
			log.Panicf("failed to migrate Program: %v", err)
		}
		return d
	})

	lago.RegistryAdmin.Register("p_nirmancampus_programs", lago.AdminPanel[Program]{
		SearchField: "Name",
		ListFields:  []string{"Name", "Code", "University"},
	})
}
