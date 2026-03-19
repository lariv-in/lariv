package p_nirmancampus_students

import (
	"log"

	"github.com/lariv-in/lago"
	"github.com/lariv-in/p_students"
	"gorm.io/gorm"
)

// NirmancampusStudentDetails is a one-to-one extension of p_students.Student.
// It stores the extra fields migrated from the Django plugin:
// - FathersName
// - Category
// - Address
type NirmancampusStudentDetails struct {
	gorm.Model

	StudentID uint `gorm:"uniqueIndex;notnull"`
	Student   p_students.Student `gorm:"constraint:OnDelete:CASCADE;foreignKey:StudentID;references:ID"`

	FathersName string `gorm:"type:varchar(255);default:''"`
	Category    string `gorm:"type:varchar(100);default:''"`
	Address     string `gorm:"type:text"`
}

func init() {
	lago.OnDBInit(func(d *gorm.DB) *gorm.DB {
		if err := d.AutoMigrate(&NirmancampusStudentDetails{}); err != nil {
			log.Panicf("failed to migrate NirmancampusStudentDetails: %v", err)
		}
		return d
	})
}

