package p_syllabus

import (
	"github.com/lariv-in/lago/lago"
	"gorm.io/gorm"
)

// SyllabusTopic flattens syllabus.Topic-style fields onto one row per course node.
type SyllabusTopic struct {
	gorm.Model

	CourseID    uint `gorm:"not null;index"`
	Title       string
	SortOrder   uint
	Book        string
	PageRange   string
	Description string `gorm:"type:text"`
	IsCompleted bool   `gorm:"not null;default:false"`
}

func init() {
	lago.OnDBInit("p_syllabus.models", func(d *gorm.DB) *gorm.DB {
		lago.RegisterModel[SyllabusTopic](d)
		return d
	})
	lago.RegistryAdmin.Register("p_syllabus", lago.AdminPanel[SyllabusTopic]{
		SearchField: "Title",
		ListFields:  []string{"CourseID", "Title", "SortOrder", "Book", "IsCompleted"},
	})
}
