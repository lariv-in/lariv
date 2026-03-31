package p_nirmancampus_programs

import (
	"log"

	"github.com/lariv-in/lago/lago"
	courses "github.com/lariv-in/lago/plugins/p_nirmancampus_courses"
	"gorm.io/gorm"
)

const (
	AdmissionSessionJan  = "jan"
	AdmissionSessionJuly = "july"
	AdmissionSessionBoth = "both"
)

const (
	TermTypeYear     = "year"
	TermTypeSemester = "semester"
)

// ProgramStructureUnit is one term of a program's structure.
// CompulsoryCourses and OptionalCourseSelectionPool are many-to-many relations to Course.
type ProgramStructureUnit struct {
	gorm.Model

	ProgramID                   uint `gorm:"not null;uniqueIndex:idx_psu_program_term"`
	TermNumber                  int  `gorm:"not null;uniqueIndex:idx_psu_program_term"`
	CompulsoryCourses           []courses.Course `gorm:"many2many:program_structure_unit_compulsory_courses;"`
	OptionalCourseCount         int
	OptionalCourseSelectionPool []courses.Course `gorm:"many2many:program_structure_unit_optional_courses;"`

	Program Program `gorm:"constraint:OnDelete:CASCADE"`
}

type Program struct {
	gorm.Model

	Name              string
	Code              string `gorm:"uniqueIndex"`
	Description       string
	University        string `gorm:"type:varchar(32);not null;default:''"`
	ProgramType       string `gorm:"type:varchar(32);not null;default:''"`
	AdmissionSessions string `gorm:"type:varchar(32);not null;default:''"`
	TermType          string `gorm:"type:varchar(32);not null;default:''"`

	ProgramStructureUnits []ProgramStructureUnit `gorm:"foreignKey:ProgramID"`
}

func init() {
	lago.OnDBInit(func(d *gorm.DB) *gorm.DB {
		if err := d.AutoMigrate(&Program{}, &ProgramStructureUnit{}); err != nil {
			log.Panicf("failed to migrate Program / ProgramStructureUnit: %v", err)
		}
		return d
	})

	lago.RegistryAdmin.Register("p_nirmancampus_programs", lago.AdminPanel[Program]{
		SearchField: "Name",
		ListFields:  []string{"Name", "Code", "University", "ProgramType"},
	})
}
