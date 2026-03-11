package p_totschool_tally

import (
	"time"

	"github.com/lariv-in/lago"
	"github.com/lariv-in/p_users"
	"gorm.io/gorm"
)

type TotSchoolSession struct {
	gorm.Model
	Name  string    `gorm:"uniqueIndex;size:250"`
	Start time.Time `gorm:"type:date"`
	End   time.Time `gorm:"type:date"`
}

func (s *TotSchoolSession) IsActive() bool {
	now := time.Now().Truncate(24 * time.Hour)
	return !s.Start.After(now) && !s.End.Before(now)
}

type Tally struct {
	gorm.Model
	UserID        uint         `gorm:"uniqueIndex:idx_user_date"`
	User          p_users.User `gorm:"foreignKey:UserID"`
	Date          time.Time    `gorm:"type:date;uniqueIndex:idx_user_date"`
	Visits        int          `gorm:"default:0"`
	Appointments  int          `gorm:"default:0"`
	Leads         int          `gorm:"default:0"`
	Presentations int          `gorm:"default:0"`
	Demos         int          `gorm:"default:0"`
	Letters       int          `gorm:"default:0"`
	FollowUps     int          `gorm:"default:0"`
	Proposals     int          `gorm:"default:0"`
	Policies      int          `json:"policies"`
	Premium       int          `json:"premium"`
}

func init() {
	lago.OnDBInit(func(d *gorm.DB) *gorm.DB {
		err := d.AutoMigrate(&TotSchoolSession{}, &Tally{})
		if err != nil {
			panic(err)
		}
		return d
	})
}
