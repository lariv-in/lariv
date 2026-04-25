package p_nirmancampus_sessions

import (
	"fmt"
	"strings"
	"time"

	"github.com/lariv-in/lago/lago"
	"gorm.io/gorm"
)

// AdmissionSession represents an admission-period window.
type AdmissionSession struct {
	gorm.Model

	Name     string
	Code     string `gorm:"uniqueIndex;default:''"`
	Start    time.Time
	End      time.Time
	IsActive bool `gorm:"default:true"`
}

// ExamSession uses the same fields for an exam-period window.
type ExamSession struct {
	gorm.Model

	Name     string
	Code     string `gorm:"uniqueIndex;default:''"`
	Start    time.Time
	End      time.Time
	IsActive bool `gorm:"default:true"`
}

func (s *AdmissionSession) BeforeSave(tx *gorm.DB) error {
	if strings.TrimSpace(s.Code) != "" || s.Start.IsZero() {
		return nil
	}
	var exclude uint
	if s.ID > 0 {
		exclude = s.ID
	}
	code, err := generateSessionMonthCode(tx, s.Start, exclude, &AdmissionSession{})
	if err != nil {
		return err
	}
	s.Code = code
	return nil
}

func (s *ExamSession) BeforeSave(tx *gorm.DB) error {
	if strings.TrimSpace(s.Code) != "" || s.Start.IsZero() {
		return nil
	}
	var exclude uint
	if s.ID > 0 {
		exclude = s.ID
	}
	code, err := generateSessionMonthCode(tx, s.Start, exclude, &ExamSession{})
	if err != nil {
		return err
	}
	s.Code = code
	return nil
}

// generateSessionMonthCode builds a code like JAN2026-1 for rows in the same calendar
// month as start, counting existing rows in the same table as model.
func generateSessionMonthCode(db *gorm.DB, start time.Time, excludeID uint, model any) (string, error) {
	if start.IsZero() {
		return "", nil
	}
	monthStart := time.Date(start.Year(), start.Month(), 1, 0, 0, 0, 0, start.Location())
	monthEnd := monthStart.AddDate(0, 1, 0)
	var count int64
	q := db.Model(model).
		Where("start >= ? AND start < ?", monthStart, monthEnd)
	if excludeID > 0 {
		q = q.Where("id <> ?", excludeID)
	}
	if err := q.Count(&count).Error; err != nil {
		return "", err
	}
	currentMonthStr := strings.ToUpper(start.Format("January"))[:4]
	return fmt.Sprintf("%s%d-%d", currentMonthStr, start.Year(), count+1), nil
}

func init() {
	lago.OnDBInit("p_nirmancampus_sessions.models", func(d *gorm.DB) *gorm.DB {
		lago.RegisterModel[AdmissionSession](d)
		lago.RegisterModel[ExamSession](d)
		return d
	})

	lago.RegistryAdmin.Register("p_nirmancampus_sessions", lago.AdminPanel[AdmissionSession]{
		SearchField: "Name",
		ListFields:  []string{"Name", "Code", "Start", "End", "IsActive"},
	})
	lago.RegistryAdmin.Register("p_nirmancampus_sessions.exam_sessions", lago.AdminPanel[ExamSession]{
		SearchField: "Name",
		ListFields:  []string{"Name", "Code", "Start", "End", "IsActive"},
	})
}
