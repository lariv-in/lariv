package p_nirmancampus_studentapplications

import (
	"log"
	"time"

	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_filesystem"
	"github.com/lariv-in/lago/plugins/p_nirmancampus_programs"
	"gorm.io/gorm"
)

// StudentApplication stores an application submitted for a student (intake record).
type StudentApplication struct {
	gorm.Model

	Name              string `gorm:"notnull"`
	ProgramID         uint   `gorm:"notnull"`
	Program           p_nirmancampus_programs.Program
	StudentName       string `gorm:"notnull"`
	DOB               *time.Time
	MotherName        string
	FatherName        string
	Category          string
	CompleteAddress   string
	Mobile            string
	PhotoID           *uint
	Photo             p_filesystem.VNode
	Documents         []p_filesystem.VNode `gorm:"many2many:student_application_documents;"`
}

func init() {
	lago.OnDBInit(func(d *gorm.DB) *gorm.DB {
		if err := d.AutoMigrate(&StudentApplication{}); err != nil {
			log.Panicf("failed to migrate StudentApplication model: %v", err)
		}
		return d
	})

	lago.RegistryAdmin.Register("p_nirmancampus_studentapplications", lago.AdminPanel[StudentApplication]{
		SearchField: "Name",
		ListFields: []string{
			"Name",
			"Program.Name",
			"StudentName",
			"DOB",
			"MotherName",
			"FatherName",
			"Category",
			"Mobile",
			"UpdatedAt",
		},
		Preload: []string{"Program"},
	})
}
