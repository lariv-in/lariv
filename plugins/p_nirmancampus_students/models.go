package p_nirmancampus_students

import (
	"log/slog"
	"time"

	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_filesystem"
	"github.com/lariv-in/lago/plugins/p_users"
	"github.com/lariv-in/lago/registry"
	"gorm.io/gorm"
)

// StudentCategoryChoices is stored category Key -> select label Value (slice order = dropdown order).
var StudentCategoryChoices = []registry.Pair[string, string]{
	{Key: "GEN", Value: "General"},
	{Key: "OBC", Value: "OBC"},
	{Key: "SC", Value: "SC"},
	{Key: "ST", Value: "ST"},
}

type Student struct {
	gorm.Model

	Name      string     `gorm:"type:varchar(255);notnull;default:''"`
	Email     string     `gorm:"type:varchar(255);default:'';index"`
	Phone     string     `gorm:"type:varchar(64);default:''"`
	StudentNo string     `gorm:"uniqueIndex;notnull"`
	DOB       *time.Time `gorm:"type:date"`

	MotherName string `gorm:"type:varchar(255);default:''"`
	FatherName string `gorm:"column:fathers_name;type:varchar(255);default:''"`
	Category   string `gorm:"type:varchar(100);default:''"`
	Address    string `gorm:"type:text"`
	PhotoID    *uint
	Photo      p_filesystem.VNode
	Documents  []p_filesystem.VNode `gorm:"many2many:student_documents;"`
}

// legacyStudentUserID is only for schema cleanup: Student no longer has UserID, but GORM
// AutoMigrate does not drop columns, so existing databases keep a NOT NULL user_id and
// inserts fail until the column is removed.
type legacyStudentUserID struct {
	UserID uint `gorm:"column:user_id"`
}

func (legacyStudentUserID) TableName() string { return "students" }

// migrateSchema applies one-off schema fixes that AutoMigrate will not perform
// (e.g. dropping columns removed from the Student model).
func migrateSchema(db *gorm.DB) {
	var stub legacyStudentUserID
	if !db.Migrator().HasTable(&stub) || !db.Migrator().HasColumn(&stub, "UserID") {
		return
	}
	if err := db.Migrator().DropColumn(&stub, "UserID"); err != nil {
		slog.Error("students: failed to drop legacy user_id column (fix DB manually: ALTER TABLE students DROP COLUMN user_id)", "error", err)
	}
}

func init() {
	lago.OnDBInit(func(d *gorm.DB) *gorm.DB {
		migrateSchema(d)
		lago.RegisterModel[Student](d)
		d.FirstOrCreate(&p_users.Role{}, p_users.Role{Name: "student"})
		return d
	})

	lago.RegistryAdmin.Register("p_nirmancampus_students", lago.AdminPanel[Student]{
		SearchField: "StudentNo",
		ListFields: []string{
			"StudentNo",
			"Name",
			"Email",
			"Phone",
			"DOB",
			"MotherName",
			"FatherName",
			"Category",
			"UpdatedAt",
		},
		Preload: []string{},
	})
}
