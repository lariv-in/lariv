package p_totschool_tally

import (
	"fmt"
	"time"

	"gorm.io/gorm"
)

// GetQuarterDetailsForDate calculates the quarter details for a given date.
func GetQuarterDetailsForDate(date time.Time) (string, time.Time, time.Time) {
	year := date.Year()
	quarter := (int(date.Month())-1)/3 + 1

	startDate := time.Date(year, time.Month((quarter-1)*3+1), 1, 0, 0, 0, 0, date.Location())

	var endDate time.Time
	if quarter == 4 {
		endDate = time.Date(year+1, 1, 1, 0, 0, 0, 0, date.Location()).Add(-24 * time.Hour)
	} else {
		endDate = time.Date(year, time.Month(quarter*3+1), 1, 0, 0, 0, 0, date.Location()).Add(-24 * time.Hour)
	}

	name := fmt.Sprintf("%d Quarter %d", year, quarter)
	return name, startDate, endDate
}

// EnsureSessionForDate ensures a TotSchoolSession exists for the given date's quarter.
func EnsureSessionForDate(db *gorm.DB, date time.Time) TotSchoolSession {
	name, startDate, endDate := GetQuarterDetailsForDate(date)

	var session TotSchoolSession
	err := db.Where("name = ?", name).First(&session).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			session = TotSchoolSession{
				Name:  name,
				Start: startDate,
				End:   endDate,
			}
			db.Create(&session)
		}
	}
	return session
}
