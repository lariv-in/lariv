package p_nirmancampus_programs

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"log"

	"github.com/lariv-in/lago/lago"
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

// CompulsoryCourses and OptionalCourseSelectionPool hold Course.Code values from p_nirmancampus_courses.
type ProgramStructureUnit struct {
	TermNumber                  int      `json:"term_number"`
	CompulsoryCourses           []string `json:"compulsory_courses"`
	OptionalCourseCount         int      `json:"optional_course_count"`
	OptionalCourseSelectionPool []string `json:"optional_course_selection_pool"`
}

// ProgramStructure is the JSONB-encoded list of ProgramStructureUnit for a program.
type ProgramStructure []ProgramStructureUnit

// Scan implements sql.Scanner for reading JSONB into ProgramStructure.
func (s *ProgramStructure) Scan(value interface{}) error {
	if value == nil {
		*s = nil
		return nil
	}
	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return fmt.Errorf("cannot scan %T into ProgramStructure", value)
	}
	if len(bytes) == 0 {
		*s = ProgramStructure{}
		return nil
	}
	var units []ProgramStructureUnit
	if err := json.Unmarshal(bytes, &units); err != nil {
		return err
	}
	*s = ProgramStructure(units)
	return nil
}

// Value implements driver.Valuer for writing ProgramStructure to JSONB.
func (s ProgramStructure) Value() (driver.Value, error) {
	if len(s) == 0 {
		return []byte("[]"), nil
	}
	b, err := json.Marshal([]ProgramStructureUnit(s))
	if err != nil {
		return nil, err
	}
	return b, nil
}

type Program struct {
	gorm.Model

	Name              string
	Code              string `gorm:"uniqueIndex"`
	Description       string
	University        string           `gorm:"type:varchar(32);not null;default:''"`
	ProgramType       string           `gorm:"type:varchar(32);not null;default:''"`
	AdmissionSessions string           `gorm:"type:varchar(32);not null;default:''"`
	TermType          string           `gorm:"type:varchar(32);not null;default:''"`
	Structure         ProgramStructure `gorm:"type:jsonb"`
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
		ListFields:  []string{"Name", "Code", "University", "ProgramType"},
	})
}
