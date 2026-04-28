package p_forums

import (
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_courses"
	"github.com/lariv-in/lago/plugins/p_users"
	"gorm.io/gorm"
)

// ForumThread mirrors forums.ForumThread (Description = Django `description`).
type ForumThread struct {
	gorm.Model

	Title       string
	Description string `gorm:"type:text"`
	CourseID    uint   `gorm:"not null;index"`
	Course      p_courses.Course
	UserID      *uint  `gorm:"index"`
	Author      *p_users.User `gorm:"foreignKey:UserID"`
	Locked      bool   `gorm:"not null;default:false"`
}

func init() {
	lago.OnDBInit("p_forums.models", func(d *gorm.DB) *gorm.DB {
		lago.RegisterModel[ForumThread](d)
		return d
	})
	lago.RegistryAdmin.Register("p_forums", lago.AdminPanel[ForumThread]{
		SearchField: "Title",
		ListFields:  []string{"Title", "CourseID", "Locked"},
	})
}
