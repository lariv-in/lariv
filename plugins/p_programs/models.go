package p_programs

import (
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_students"
	"github.com/lariv-in/lago/plugins/p_teachers"
	"gorm.io/gorm"
)

// Program mirrors allocation.Batch (batch / program template).
// Students / Teachers = Django M2M on Batch.
type Program struct {
	gorm.Model

	Code        string `gorm:"not null;uniqueIndex:ux_program_semester_code"`
	Name        string `gorm:"not null"`
	Standard    string
	Description string `gorm:"type:text"`
	IsActive    bool   `gorm:"not null;default:true"`
	Fee         uint   `gorm:"not null;default:0"`
	SemesterID  *uint  `gorm:"index;uniqueIndex:ux_program_semester_code"`

	Students []p_students.Student `gorm:"many2many:program_students;"`
	Teachers []p_teachers.Teacher `gorm:"many2many:program_teachers;"`
}

func init() {
	lago.OnDBInit("p_programs.models", func(d *gorm.DB) *gorm.DB {
		lago.RegisterModel[Program](d)
		return d
	})
	lago.RegistryAdmin.Register("p_programs", lago.AdminPanel[Program]{
		SearchField: "Name",
		ListFields:  []string{"Code", "Name", "Standard", "IsActive", "SemesterID"},
	})
}
