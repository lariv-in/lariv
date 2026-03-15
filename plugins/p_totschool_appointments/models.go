package p_totschool_appointments

import (
	"time"

	"github.com/lariv-in/lago"
	"github.com/lariv-in/p_users"
	"gorm.io/gorm"
)

type Appointment struct {
	gorm.Model
	CreatedByID     uint         `gorm:"notnull"`
	CreatedBy       p_users.User `gorm:"foreignKey:CreatedByID"`
	Name            string       `gorm:"size:250;notnull"`
	Location        string       `gorm:"type:text"`
	Datetime        time.Time    `gorm:"notnull"`
	Phone           string       `gorm:"size:20"`
	Remarks         string       `gorm:"type:text"`
	ExtraInfo       string       `gorm:"type:text"`
	GeneratedLetter string       `gorm:"type:text"`
	GenerationID    *int         // non-nil while AI generation is in progress
}

func (a *Appointment) GetOverlappingAppointments(db *gorm.DB) []Appointment {
	if a.CreatedByID == 0 || a.Datetime.IsZero() {
		return nil
	}
	var results []Appointment
	db.Where("created_by_id = ? AND datetime >= ? AND datetime <= ? AND id != ?",
		a.CreatedByID,
		a.Datetime.Add(-30*time.Minute),
		a.Datetime.Add(30*time.Minute),
		a.ID,
	).Find(&results)
	return results
}

func init() {
	lago.OnDBInit(func(d *gorm.DB) *gorm.DB {
		if err := d.AutoMigrate(&Appointment{}); err != nil {
			panic(err)
		}
		d.Model(&Appointment{}).Where("generation_id IS NOT NULL").Update("generation_id", nil)
		go runWorker(d)
		return d
	})
	lago.RegistryAdmin.Register("p_totschool_appointments.Appointment", lago.AdminPanel[Appointment]{SearchField: "Name"})
}
