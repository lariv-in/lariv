package p_assessments

import (
	"time"

	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_syllabus"
	"gorm.io/gorm"
)

// Assessment mirrors assessments.Assessment (exam definition; UI can be added later).
type Assessment struct {
	gorm.Model

	Title        string
	Description  string `gorm:"type:text"`
	IsActive     bool   `gorm:"not null;default:true"`
	CreatedByID  *uint  `gorm:"index"`
	Syllabus     string `gorm:"type:text"`
	WhenAt       time.Time
	Venue        string
	CourseID     *uint `gorm:"index"`
	SemesterID   *uint `gorm:"index"`
	MaxMarks     int
	PassingMarks int

	// Topics = Django assessments.Assessment.topics (syllabus.Topic); join table for GORM M2M.
	Topics []p_syllabus.SyllabusTopic `gorm:"many2many:assessment_topics;"`
}

// GradeEntry is kept for Sarathi component scores; also carries Django-like remarks/status on submissions.
type GradeEntry struct {
	gorm.Model

	StudentID uint `gorm:"not null;index"`
	CourseID  uint `gorm:"not null;index"`
	Component string
	Score     float64
	MaxScore  float64
	Remarks   string `gorm:"type:text"`
	Status    string // e.g. PASS / FAIL when mirroring StudentAssessment
}

func init() {
	lago.OnDBInit("p_assessments.models", func(d *gorm.DB) *gorm.DB {
		lago.RegisterModel[Assessment](d)
		lago.RegisterModel[GradeEntry](d)
		return d
	})
	lago.RegistryAdmin.Register("p_assessments", lago.AdminPanel[GradeEntry]{
		SearchField: "Component",
		ListFields:  []string{"StudentID", "CourseID", "Component", "Score", "MaxScore", "Status"},
	})
}
