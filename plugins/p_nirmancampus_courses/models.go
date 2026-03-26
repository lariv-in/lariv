package p_nirmancampus_courses

import (
	"log"

	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_nirmancampus_programs"
	"gorm.io/gorm"
)

type Course struct {
	gorm.Model

	Name        string
	IsActive    bool
	Code        string
	Subject     string
	Description string
}

type CourseProgram struct {
	gorm.Model

	CourseID uint   `gorm:"not null;index;uniqueIndex:idx_course_program_pair"`
	Course   Course `gorm:"constraint:OnDelete:CASCADE;foreignKey:CourseID;references:ID"`

	ProgramID uint                            `gorm:"not null;index;uniqueIndex:idx_course_program_pair"`
	Program   p_nirmancampus_programs.Program `gorm:"constraint:OnDelete:CASCADE;foreignKey:ProgramID;references:ID"`

	Semester uint `gorm:"not null;index"`
}

func init() {
	lago.OnDBInit(func(d *gorm.DB) *gorm.DB {
		if err := d.AutoMigrate(&Course{}, &CourseProgram{}); err != nil {
			log.Panicf("failed to migrate Course/CourseProgram models: %v", err)
		}
		return d
	})
	lago.RegistryAdmin.Register("p_nirmancampus_courses", lago.AdminPanel[Course]{SearchField: "Name"})
	lago.RegistryAdmin.Register("p_nirmancampus_courseprograms", lago.AdminPanel[CourseProgram]{
		SearchField: "Semester",
		ListFields:  []string{"Course.Name", "Program.Name", "Semester", "UpdatedAt"},
		Preload:     []string{"Course", "Program"},
	})
}
