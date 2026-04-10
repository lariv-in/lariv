package p_nirmancampus_sessions

import (
	"fmt"
	"strings"
	"time"

	"github.com/lariv-in/lago/lago"
	"gorm.io/gorm"
)

// Session represents an academic session window.
//
// Source parity notes (Django):
// - code is optional on create, auto-generated on save
// - code is unique
type Session struct {
	gorm.Model

	Name     string
	Code     string `gorm:"uniqueIndex;default:''"`
	Start    time.Time
	End      time.Time
	IsActive bool `gorm:"default:true"`
}

func (s *Session) BeforeSave(tx *gorm.DB) error {
	if strings.TrimSpace(s.Code) != "" || s.Start.IsZero() {
		return nil
	}

	// Generate code grouped by (start month, start year), matching the Django logic.
	monthStart := time.Date(s.Start.Year(), s.Start.Month(), 1, 0, 0, 0, 0, s.Start.Location())
	monthEnd := monthStart.AddDate(0, 1, 0)

	var count int64
	if err := tx.Model(&Session{}).
		Where("start >= ? AND start < ?", monthStart, monthEnd).
		Where("id <> ?", s.ID).
		Count(&count).Error; err != nil {
		return err
	}

	currentMonthStr := strings.ToUpper(s.Start.Format("January"))[:4]
	s.Code = fmt.Sprintf("%s%d-%d", currentMonthStr, s.Start.Year(), count+1)
	return nil
}

func init() {
	lago.OnDBInit("p_nirmancampus_sessions.models", func(d *gorm.DB) *gorm.DB {
		lago.RegisterModel[Session](d)
		return d
	})

	lago.RegistryAdmin.Register("p_nirmancampus_sessions", lago.AdminPanel[Session]{
		SearchField: "Name",
		ListFields:  []string{"Name", "Code", "Start", "End", "IsActive"},
	})
}
