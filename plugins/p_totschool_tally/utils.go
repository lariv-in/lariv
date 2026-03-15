package p_totschool_tally

import (
	"context"
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

func CurrentSessionNameForDateGetter(ctx context.Context) any {
	db := ctx.Value("$db").(*gorm.DB)
	date := time.Now()
	session := EnsureSessionForDate(db, date)
	return session.Name
}

// FormatCurrencyIndian formats an integer amount using the Indian numbering system,
// e.g. 1234567 -> "₹12,34,567".
func FormatCurrencyIndian(amount int) string {
	if amount == 0 {
		return "₹0"
	}

	s := fmt.Sprintf("%d", amount)
	if len(s) <= 3 {
		return "₹" + s
	}

	result := s[len(s)-3:]
	s = s[:len(s)-3]

	for len(s) > 0 {
		if len(s) <= 2 {
			result = s + "," + result
			break
		}
		result = s[len(s)-2:] + "," + result
		s = s[:len(s)-2]
	}

	return "₹" + result
}

// BuildWhatsappMessage constructs the WhatsApp report message text
// mirroring the original Django implementation.
func BuildWhatsappMessage(data WhatsappReportData) string {
	if !data.Submitted {
		return ""
	}

	dateStr := data.Date.Format("02 Jan, 2006")

	message := "TOT School Report\n"
	message += fmt.Sprintf("Date: %s\n", dateStr)
	message += fmt.Sprintf("Name: %s\n\n", data.UserName)

	// Today / QTD / Last quarter metrics
	message += fmt.Sprintf("- Visits: %d/%d/%d\n", data.Today.TotalVisits, data.QTD.TotalVisits, data.LastQuarter.TotalVisits)
	message += fmt.Sprintf("- Appointments: %d/%d/%d\n", data.Today.TotalAppointments, data.QTD.TotalAppointments, data.LastQuarter.TotalAppointments)
	message += fmt.Sprintf("- Leads: %d/%d/%d\n", data.Today.TotalLeads, data.QTD.TotalLeads, data.LastQuarter.TotalLeads)
	message += fmt.Sprintf("- Presentations: %d/%d/%d\n", data.Today.TotalPresentations, data.QTD.TotalPresentations, data.LastQuarter.TotalPresentations)
	message += fmt.Sprintf("- Demonstrations: %d/%d/%d\n", data.Today.TotalDemos, data.QTD.TotalDemos, data.LastQuarter.TotalDemos)
	message += fmt.Sprintf("- Follow Up Letters: %d/%d/%d\n", data.Today.TotalLetters, data.QTD.TotalLetters, data.LastQuarter.TotalLetters)
	message += fmt.Sprintf("- Follow Ups: %d/%d/%d\n", data.Today.TotalFollowUps, data.QTD.TotalFollowUps, data.LastQuarter.TotalFollowUps)
	message += fmt.Sprintf("- Proposals Given: %d/%d/%d\n", data.Today.TotalProposals, data.QTD.TotalProposals, data.LastQuarter.TotalProposals)

	// Premium uses Indian currency formatting
	message += fmt.Sprintf(
		"- Premium: %s/%s/%s\n",
		FormatCurrencyIndian(data.Today.TotalPremium),
		FormatCurrencyIndian(data.QTD.TotalPremium),
		FormatCurrencyIndian(data.LastQuarter.TotalPremium),
	)

	return message
}
