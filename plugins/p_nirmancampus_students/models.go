package p_nirmancampus_students

import (
	"time"

	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_filesystem"
	"github.com/lariv-in/lago/plugins/p_users"
	"gorm.io/gorm"
)

type Student struct {
	gorm.Model

	UserID    uint         `gorm:"uniqueIndex;notnull"`
	User      p_users.User `gorm:"constraint:OnDelete:CASCADE"`
	StudentNo string       `gorm:"uniqueIndex;notnull"`
	DOB       *time.Time   `gorm:"type:date"`

	MotherName string `gorm:"type:varchar(255);default:''"`
	FatherName string `gorm:"column:fathers_name;type:varchar(255);default:''"`
	Category   string `gorm:"type:varchar(100);default:''"`
	Address    string `gorm:"type:text"`
	PhotoID    *uint
	Photo      p_filesystem.VNode
	Documents  []p_filesystem.VNode `gorm:"many2many:student_documents;"`
}

func init() {
	lago.OnDBInit(func(d *gorm.DB) *gorm.DB {
		lago.RegisterModel[Student](d)
		d.FirstOrCreate(&p_users.Role{}, p_users.Role{Name: "student"})
		return d
	})

	lago.RegistryAdmin.Register("p_nirmancampus_students", lago.AdminPanel[Student]{
		SearchField: "StudentNo",
		ListFields: []string{
			"StudentNo",
			"User.Name",
			"User.Email",
			"User.Phone",
			"DOB",
			"MotherName",
			"FatherName",
			"Category",
			"UpdatedAt",
		},
		Preload: []string{"User"},
	})
}
