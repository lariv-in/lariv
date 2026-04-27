package p_finances

import (
	"time"

	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_semesters"
	"github.com/lariv-in/lago/plugins/p_students"
	"gorm.io/gorm"
)

// StudentCharge is a fee line; extra columns align with finances.Transaction purpose/semester/remarks.
type StudentCharge struct {
	gorm.Model

	StudentID   uint `gorm:"not null;index"`
	Student     p_students.Student
	AmountCents int64
	Description string // maps to Django transaction purpose / line description
	Purpose     string
	Remarks     string `gorm:"type:text"`
	DueOn       *time.Time
	SemesterID  *uint `gorm:"index"`
	Semester    *p_semesters.Semester `gorm:"foreignKey:SemesterID"`
}

func init() {
	lago.OnDBInit("p_finances.models", func(d *gorm.DB) *gorm.DB {
		lago.RegisterModel[StudentCharge](d)
		return d
	})
	lago.RegistryAdmin.Register("p_finances", lago.AdminPanel[StudentCharge]{
		SearchField: "Description",
		ListFields:  []string{"StudentID", "AmountCents", "Description", "Purpose", "SemesterID", "DueOn"},
	})
}
