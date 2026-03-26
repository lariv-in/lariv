package p_nirmancampus_students

import (
	"log"
	"time"

	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_filesystem"
	"github.com/lariv-in/lago/plugins/p_users"
	"gorm.io/gorm"
)

// Student is the single students row for Nirmancampus (replaces p_students.Student + NirmancampusStudentDetails).
//
// If you already had nirmancampus_student_details, after AutoMigrate adds columns on students (PostgreSQL):
//
//	UPDATE students AS s
//	SET
//	  fathers_name = d.fathers_name,
//	  category = d.category,
//	  address = d.address
//	FROM nirmancampus_student_details AS d
//	WHERE d.student_id = s.id AND d.deleted_at IS NULL;
//	DROP TABLE nirmancampus_student_details;
//
// Confirm table/column names in your database before running.
type Student struct {
	gorm.Model

	UserID    uint         `gorm:"uniqueIndex;notnull"`
	User      p_users.User `gorm:"constraint:OnDelete:CASCADE"`
	StudentNo string       `gorm:"uniqueIndex;notnull"`
	DOB       *time.Time

	Assets []p_filesystem.VNode `gorm:"many2many:student_assets;"`

	FathersName string `gorm:"type:varchar(255);default:''"`
	Category    string `gorm:"type:varchar(100);default:''"`
	Address     string `gorm:"type:text"`
}

func init() {
	lago.OnDBInit(func(d *gorm.DB) *gorm.DB {
		if err := d.AutoMigrate(&Student{}); err != nil {
			log.Panicf("failed to migrate Student model: %v", err)
		}
		d.FirstOrCreate(&p_users.Role{}, p_users.Role{Name: "student"})
		return d
	})

	lago.RegistryAdmin.Register("p_nirmancampus_students", lago.AdminPanel[Student]{
		SearchField: "StudentNo",
		ListFields:  []string{"StudentNo", "User.Name", "DOB", "FathersName", "Category", "UpdatedAt"},
		Preload:     []string{"User"},
	})
}
