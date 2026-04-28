package p_allocation

import (
	"github.com/lariv-in/lago/lago"
	"gorm.io/gorm"
)

// CourseTeacherAssignment links a teacher to a course offering (IDs only; no cross-plugin GORM associations).
type CourseTeacherAssignment struct {
	gorm.Model

	TeacherID uint `gorm:"not null;index"`
	CourseID  uint `gorm:"not null;index"`
	Role      string
}

func init() {
	lago.OnDBInit("p_allocation.models", func(d *gorm.DB) *gorm.DB {
		lago.RegisterModel[CourseTeacherAssignment](d)
		return d
	})
	lago.RegistryAdmin.Register("p_allocation", lago.AdminPanel[CourseTeacherAssignment]{
		SearchField: "Role",
		ListFields:  []string{"TeacherID", "CourseID", "Role"},
	})
}
