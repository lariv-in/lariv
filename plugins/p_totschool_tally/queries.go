package p_totschool_tally

import (
	"fmt"
	"sort"
	"time"

	"github.com/lariv-in/p_users"
	"gorm.io/gorm"
)

type DashboardStats struct {
	TotalPresentations int
	TotalLeads         int
	TotalVisits        int
	TotalAppointments  int
	TotalDemos         int
	TotalLetters       int
	TotalFollowUps     int
	TotalProposals     int
	TotalPolicies      int
	TotalPremium       int
	FormsFilled        int
	ApptVisitRatio     float64
	DemoApptRatio      float64
	PolicyDemoRatio    float64
}

func GetDashboardStats(db *gorm.DB, userID *uint, session *TotSchoolSession) DashboardStats {
	query := db.Model(&Tally{})
	if userID != nil {
		query = query.Where("user_id = ?", *userID)
	}
	if session != nil {
		query = query.Where("date >= ? AND date <= ?", session.Start, session.End)
	}

	type Result struct {
		TotalPresentations int
		TotalLeads         int
		TotalVisits        int
		TotalAppointments  int
		TotalDemos         int
		TotalLetters       int
		TotalFollowUps     int
		TotalProposals     int
		TotalPolicies      int
		TotalPremium       int
		FormsFilled        int
	}
	var res Result
	query.Select(`
		COALESCE(SUM(presentations), 0) as total_presentations,
		COALESCE(SUM(leads), 0) as total_leads,
		COALESCE(SUM(visits), 0) as total_visits,
		COALESCE(SUM(appointments), 0) as total_appointments,
		COALESCE(SUM(demos), 0) as total_demos,
		COALESCE(SUM(letters), 0) as total_letters,
		COALESCE(SUM(follow_ups), 0) as total_follow_ups,
		COALESCE(SUM(proposals), 0) as total_proposals,
		COALESCE(SUM(policies), 0) as total_policies,
		COALESCE(SUM(premium), 0) as total_premium,
		COUNT(id) as forms_filled
	`).Scan(&res)

	stats := DashboardStats{
		TotalPresentations: res.TotalPresentations,
		TotalLeads:         res.TotalLeads,
		TotalVisits:        res.TotalVisits,
		TotalAppointments:  res.TotalAppointments,
		TotalDemos:         res.TotalDemos,
		TotalLetters:       res.TotalLetters,
		TotalFollowUps:     res.TotalFollowUps,
		TotalProposals:     res.TotalProposals,
		TotalPolicies:      res.TotalPolicies,
		TotalPremium:       res.TotalPremium,
		FormsFilled:        res.FormsFilled,
	}

	if res.TotalVisits > 0 {
		stats.ApptVisitRatio = float64(res.TotalAppointments) / float64(res.TotalVisits) * 100
	}
	if res.TotalAppointments > 0 {
		stats.DemoApptRatio = float64(res.TotalDemos) / float64(res.TotalAppointments) * 100
	}
	if res.TotalDemos > 0 {
		stats.PolicyDemoRatio = float64(res.TotalPolicies) / float64(res.TotalDemos) * 100
	}

	return stats
}

type WhatsappReportData struct {
	Submitted   bool
	Today       DashboardStats
	QTD         DashboardStats
	LastQuarter DashboardStats
	UserName    string
	Date        time.Time
}

func GetWhatsappReportData(db *gorm.DB, userID uint) WhatsappReportData {
	today := time.Now().Truncate(24 * time.Hour)
	var count int64
	db.Model(&Tally{}).Where("user_id = ? AND date = ?", userID, today).Count(&count)
	if count == 0 {
		return WhatsappReportData{Submitted: false}
	}

	todaySession := TotSchoolSession{Start: today, End: today}
	todayTotals := GetDashboardStats(db, &userID, &todaySession)

	currentQuarter := EnsureSessionForDate(db, today)
	qtdSession := TotSchoolSession{Start: currentQuarter.Start, End: today}
	qtdTotals := GetDashboardStats(db, &userID, &qtdSession)

	lastQuarterDate := currentQuarter.Start.Add(-24 * time.Hour)
	lastQuarterSession := EnsureSessionForDate(db, lastQuarterDate)
	lastQuarterTotals := GetDashboardStats(db, &userID, &lastQuarterSession)

	var user p_users.User
	db.First(&user, userID)

	return WhatsappReportData{
		Submitted:   true,
		Today:       todayTotals,
		QTD:         qtdTotals,
		LastQuarter: lastQuarterTotals,
		UserName:    user.Name,
		Date:        today,
	}
}

type LeaderboardEntry struct {
	Rank     string
	UserID   uint
	UserName string
	Value    int
}

type LeaderboardResult struct {
	Top5        []LeaderboardEntry
	CurrentUser *LeaderboardEntry
}

func GetLeaderboards(db *gorm.DB, userID *uint, session *TotSchoolSession) map[string]LeaderboardResult {
	query := db.Model(&Tally{}).
		Joins("JOIN users ON users.id = tallies.user_id").
		Select(`
			tallies.user_id,
			users.name as user_name,
			COALESCE(SUM(visits), 0) as total_visits,
			COALESCE(SUM(demos), 0) as total_demos,
			COALESCE(SUM(policies), 0) as total_policies,
			COALESCE(SUM(premium), 0) as total_premium
		`).
		Group("tallies.user_id, users.name")

	if session != nil {
		query = query.Where("tallies.date >= ? AND tallies.date <= ?", session.Start, session.End)
	}

	type UserTotal struct {
		UserID        uint
		UserName      string
		TotalVisits   int
		TotalDemos    int
		TotalPolicies int
		TotalPremium  int
	}

	var userTotals []UserTotal
	query.Find(&userTotals)

	metrics := []struct {
		Name  string
		Value func(UserTotal) int
	}{
		{"visits", func(u UserTotal) int { return u.TotalVisits }},
		{"demos", func(u UserTotal) int { return u.TotalDemos }},
		{"policies", func(u UserTotal) int { return u.TotalPolicies }},
		{"premium", func(u UserTotal) int { return u.TotalPremium }},
	}

	leaderboards := make(map[string]LeaderboardResult)

	for _, metric := range metrics {
		sortedTotals := make([]UserTotal, len(userTotals))
		copy(sortedTotals, userTotals)

		sort.SliceStable(sortedTotals, func(i, j int) bool {
			return metric.Value(sortedTotals[i]) > metric.Value(sortedTotals[j])
		})

		var top5 []LeaderboardEntry
		var userEntry *LeaderboardEntry

		for index, row := range sortedTotals {
			rank := index + 1
			entry := LeaderboardEntry{
				Rank:     fmt.Sprintf("%d", rank),
				UserID:   row.UserID,
				UserName: row.UserName,
				Value:    metric.Value(row),
			}

			if rank <= 5 {
				top5 = append(top5, entry)
			}

			if userID != nil && row.UserID == *userID {
				userEntry = &entry
			}
		}

		if userID != nil && userEntry == nil {
			var user p_users.User
			if err := db.First(&user, *userID).Error; err == nil {
				userEntry = &LeaderboardEntry{
					Rank:     "-",
					UserID:   user.ID,
					UserName: user.Name,
					Value:    0,
				}
			}
		}

		leaderboards[metric.Name] = LeaderboardResult{
			Top5:        top5,
			CurrentUser: userEntry,
		}
	}

	return leaderboards
}
