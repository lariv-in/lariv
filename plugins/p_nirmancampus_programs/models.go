package p_nirmancampus_programs

import (
	"log/slog"

	"github.com/lariv-in/lago/lago"
	courses "github.com/lariv-in/lago/plugins/p_nirmancampus_courses"
	"github.com/lariv-in/lago/registry"
	"gorm.io/gorm"
)

// admissionSessionChoices: stored AdmissionSessions Key -> label (slice order = dropdown order).
var admissionSessionChoices = []registry.Pair[string, string]{
	{Key: "both", Value: "January and July"},
	{Key: "jan", Value: "January"},
	{Key: "july", Value: "July"},
}

// termTypeChoices: stored TermType Key -> label.
var termTypeChoices = []registry.Pair[string, string]{
	{Key: "semester", Value: "Semester"},
	{Key: "year", Value: "Year"},
}

// ProgramStructureUnit is one term of a program's structure.
// CompulsoryCourses and OptionalCourseSelectionPool are many-to-many relations to Course.
type ProgramStructureUnit struct {
	gorm.Model

	ProgramID                   uint             `gorm:"not null;uniqueIndex:idx_psu_program_term"`
	TermNumber                  uint             `gorm:"not null;uniqueIndex:idx_psu_program_term"`
	CompulsoryCourses           []courses.Course `gorm:"many2many:program_structure_unit_compulsory_courses;"`
	OptionalCourseCount         uint
	OptionalCourseSelectionPool []courses.Course `gorm:"many2many:program_structure_unit_optional_courses;"`

	Program Program `gorm:"constraint:OnDelete:CASCADE"`
}

// ProgramMedia is a language label attachable to programs (many-to-many).
type ProgramMedia struct {
	gorm.Model
	Language string `gorm:"not null"`
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
	Fee               uint   `gorm:"not null;default:0"`

	ProgramMedia          []ProgramMedia         `gorm:"many2many:program_program_media;"`
	ProgramStructureUnits []ProgramStructureUnit `gorm:"foreignKey:ProgramID"`
}

// UniversityChoices maps stored Program.University keys to UI labels (slice order = dropdown order).
var UniversityChoices = []registry.Pair[string, string]{
	{Key: "IGNOU", Value: "IGNOU"},
	{Key: "MRSPTU", Value: "MRSPTU"},
}

var programTypeChoices = []registry.Pair[string, string]{
	{Key: "bachelor", Value: "Bachelor"},
	{Key: "certificate", Value: "Certificate"},
	{Key: "diploma", Value: "Diploma"},
	{Key: "masters", Value: "Masters"},
}

func init() {
	lago.OnDBInit("p_nirmancampus_programs.models", func(d *gorm.DB) *gorm.DB {
		lago.RegisterModel[ProgramMedia](d)
		lago.RegisterModel[Program](d)
		lago.RegisterModel[ProgramStructureUnit](d)
		for _, lang := range []string{"Hindi", "English", "Punjabi"} {
			res := d.FirstOrCreate(&ProgramMedia{}, ProgramMedia{Language: lang})
			if res.Error != nil {
				slog.Error("seed program_media", "language", lang, "error", res.Error)
			}
		}
		return d
	})

	lago.RegistryAdmin.Register("p_nirmancampus_programs", lago.AdminPanel[Program]{
		SearchField: "Name",
		ListFields:  []string{"Name", "Code", "University", "ProgramType", "Fee"},
	})
}
