package p_assignments

import (
	"time"

	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_courses"
	"github.com/lariv-in/lago/plugins/p_semesters"
	"github.com/lariv-in/lago/registry"
	"gorm.io/gorm"
)

// Assignment mirrors assignments.Assignment (Django `type` → AssignmentType column).
type Assignment struct {
	gorm.Model

	CourseID       uint `gorm:"not null;index"`
	Course         p_courses.Course
	Title          string
	Description    string `gorm:"type:text"`
	ReleaseAt      *time.Time
	DueAt          *time.Time
	CreatedByID    *uint `gorm:"index"`
	TotalMarks     int
	SemesterID     *uint  `gorm:"index"`
	Semester       *p_semesters.Semester `gorm:"foreignKey:SemesterID"`
	AssignmentType string `gorm:"column:assignment_type;type:varchar(20);default:'Online'"` // Online / Offline
}

// AssignmentTypeChoices for persisted AssignmentType.
var AssignmentTypeChoices = []registry.Pair[string, string]{
	{Key: "Online", Value: "Online"},
	{Key: "Offline", Value: "Offline"},
}

func init() {
	lago.OnDBInit("p_assignments.models", func(d *gorm.DB) *gorm.DB {
		lago.RegisterModel[Assignment](d)
		return d
	})
	lago.RegistryAdmin.Register("p_assignments", lago.AdminPanel[Assignment]{
		SearchField: "Title",
		ListFields:  []string{"CourseID", "Title", "DueAt", "TotalMarks", "AssignmentType", "SemesterID"},
	})
}
