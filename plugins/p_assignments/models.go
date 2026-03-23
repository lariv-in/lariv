package p_assignments

import (
	"log"
	"time"

	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/p_filesystem"
	"github.com/lariv-in/lago/p_semesters"
	"gorm.io/gorm"
)

// Assignment is a coursework item with a due datetime, optional attachments, and a maximum score.
type Assignment struct {
	gorm.Model

	Name        string    `gorm:"notnull"`
	Due         time.Time `gorm:"notnull"`
	Description string    `gorm:"type:text"`
	MaxMarks    int       `gorm:"notnull"`

	// Semester is non-null in Lago UI: the semester selector expects a concrete FK value.
	SemesterID uint
	Semester   p_semesters.Semester `gorm:"constraint:OnDelete:CASCADE;foreignKey:SemesterID;references:ID"`

	Assets []p_filesystem.VNode `gorm:"many2many:assignment_assets;"`
}

func init() {
	lago.OnDBInit(func(d *gorm.DB) *gorm.DB {
		if err := d.AutoMigrate(&Assignment{}); err != nil {
			log.Panicf("failed to migrate Assignment model: %v", err)
		}
		return d
	})

	lago.RegistryAdmin.Register("p_assignments", lago.AdminPanel[Assignment]{
		SearchField: "Name",
		ListFields: []string{
			"Name",
			"Due",
			"MaxMarks",
			"Semester.Name",
			"Description",
			"UpdatedAt",
		},
		Preload: []string{"Semester", "Assets"},
	})
}
