package p_teachers

import (
	"github.com/lariv-in/lago/lago"
	"gorm.io/gorm"
)

// Teacher mirrors accounts.Teacher (Code = Django `code`; Name/Email/Phone denormalized for Sarathi UI without users row).
type Teacher struct {
	gorm.Model

	Code           string `gorm:"uniqueIndex;not null"`
	Name           string `gorm:"not null"`
	Email          string
	Phone          string
	Qualifications string `gorm:"type:text"`

	UserID         *uint `gorm:"index"`
	ProfilePhotoID *uint
}

func init() {
	lago.OnDBInit("p_teachers.models", func(d *gorm.DB) *gorm.DB {
		lago.RegisterModel[Teacher](d)
		return d
	})
	lago.RegistryAdmin.Register("p_teachers", lago.AdminPanel[Teacher]{
		SearchField: "Name",
		ListFields:  []string{"Code", "Name", "Email", "Phone"},
	})
}
