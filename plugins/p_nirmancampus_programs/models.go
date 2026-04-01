package p_nirmancampus_programs

import (
	"log"

	"github.com/lariv-in/lago/lago"
	courses "github.com/lariv-in/lago/plugins/p_nirmancampus_courses"
	"github.com/lariv-in/lago/registry"
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
	TermNumber                  uint `gorm:"not null;uniqueIndex:idx_psu_program_term"`
	CompulsoryCourses           []courses.Course `gorm:"many2many:program_structure_unit_compulsory_courses;"`
	OptionalCourseCount         uint
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

func universityChoices() []registry.Pair[string, string] {
	return []registry.Pair[string, string]{
		{Key: "IGNOU", Value: "IGNOU"},
		{Key: "MRSPTU", Value: "MRSPTU"},
	}
}

func universityPairForKey(stored string) registry.Pair[string, string] {
	if stored == "" {
		return registry.Pair[string, string]{}
	}
	for _, p := range universityChoices() {
		if p.Key == stored {
			return p
		}
	}
	return registry.Pair[string, string]{Key: stored, Value: stored}
}

func programTypeChoices() []registry.Pair[string, string] {
	return []registry.Pair[string, string]{
		{Key: "certificate", Value: "Certificate"},
		{Key: "diploma", Value: "Diploma"},
		{Key: "bachelor", Value: "Bachelor"},
		{Key: "masters", Value: "Masters"},
	}
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
