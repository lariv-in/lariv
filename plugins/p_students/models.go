package p_students

import (
	"time"

	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_filesystem"
	"gorm.io/gorm"
)

// Student mirrors accounts.Student (single-institute Sarathi; Name/Email/Phone denormalized when no User row).
// Documents = Django M2M attachments on student (filesystem file nodes).
type Student struct {
	gorm.Model

	StudentNo string `gorm:"uniqueIndex;not null"`
	Name      string `gorm:"not null"`
	Email     string
	Phone     string

	UserID         *uint `gorm:"index"`
	ProfilePhotoID *uint

	AdhaarNo     string
	DOB          *time.Time
	Gender       string
	Nationality  string
	MotherTongue string
	Religion     string
	Caste        string
	Category     string
	SpecialNeeds string `gorm:"type:text"`
	Address      string `gorm:"type:text"`

	PrevSchoolName      string
	PrevSchoolAddress   string `gorm:"type:text"`
	PrevSchoolClass     string
	PrevSchoolPassDate  *time.Time
	PrevSchoolUDISECode string

	Guardian1Name  string
	Guardian1Email string
	Guardian1Phone string
	Guardian2Name  string
	Guardian2Email string
	Guardian2Phone string

	Documents []p_filesystem.VNode `gorm:"many2many:student_documents;"`
}

func init() {
	lago.OnDBInit("p_students.models", func(d *gorm.DB) *gorm.DB {
		lago.RegisterModel[Student](d)
		return d
	})
	lago.RegistryAdmin.Register("p_students", lago.AdminPanel[Student]{
		SearchField: "Name",
		ListFields: []string{
			"StudentNo", "Name", "Email", "Phone",
			"Gender", "Nationality", "Guardian1Name", "Guardian1Phone",
		},
	})
}
