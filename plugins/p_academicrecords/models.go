package p_academicrecords

import (
	"log"

	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_nirmancampus_students"
	"gorm.io/gorm"
)

// AcademicRecord links a Student with a simple status.
type AcademicRecord struct {
	gorm.Model

	StudentID uint `gorm:"notnull;index"`
	Student   p_nirmancampus_students.Student `gorm:"constraint:OnDelete:CASCADE;foreignKey:StudentID;references:ID"`

	Status string `gorm:"type:varchar(50);notnull"`
}

func init() {
	lago.OnDBInit(func(d *gorm.DB) *gorm.DB {
		if err := d.AutoMigrate(&AcademicRecord{}); err != nil {
			log.Panicf("failed to migrate AcademicRecord: %v", err)
		}
		return d
	})

	lago.RegistryAdmin.Register("p_academicrecords", lago.AdminPanel[AcademicRecord]{
		SearchField: "Status",
		ListFields:  []string{"Status", "Student.StudentNo", "UpdatedAt"},
		Preload:     []string{"Student.User"},
	})
}

