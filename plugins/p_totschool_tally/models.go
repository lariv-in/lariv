package p_totschool_tally

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_users"
	"gorm.io/gorm"
)

type TotSchoolSession struct {
	gorm.Model
	// Use `unique` (not `uniqueIndex`): with uniqueIndex, PostgreSQL reports the
	// column as unique but GORM's field.Unique stays false, so AutoMigrate tries
	// ALTER TABLE ... DROP CONSTRAINT uni_* — which fails because uniqueness is
	// enforced by an index, not that constraint name.
	Name  string    `gorm:"unique;size:250"`
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

func (t *Tally) BeforeSave(tx *gorm.DB) (err error) {
	EnsureSessionForDate(tx, t.Date)
	return nil
}

// EnsureSessionForDate ensures a TotSchoolSession exists for the given date's quarter.
func EnsureSessionForDate(db *gorm.DB, date time.Time) TotSchoolSession {
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

	session, err := gorm.G[TotSchoolSession](db).Where("name = ?", name).First(context.Background())
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			session = TotSchoolSession{
				Name:  name,
				Start: startDate,
				End:   endDate,
			}
			_ = gorm.G[TotSchoolSession](db).Create(context.Background(), &session)
		}
	}
	return session
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

func GetWhatsappReportData(db *gorm.DB, userID uint) WhatsappReportData {
	today := time.Now().Truncate(24 * time.Hour)
	count, err := gorm.G[Tally](db).Where("user_id = ? AND date = ?", userID, today).Count(context.Background(), "*")
	if err != nil || count == 0 {
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

	user, _ := gorm.G[p_users.User](db).Where("id = ?", userID).First(context.Background())

	return WhatsappReportData{
		Submitted:   true,
		Today:       todayTotals,
		QTD:         qtdTotals,
		LastQuarter: lastQuarterTotals,
		UserName:    user.Name,
		Date:        today,
	}
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
			user, err := gorm.G[p_users.User](db).Where("id = ?", *userID).First(context.Background())
			if err == nil {
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

func init() {
	lago.OnDBInit("p_totschool_tally.models", func(d *gorm.DB) *gorm.DB {
		lago.RegisterModel[TotSchoolSession](d)
		lago.RegisterModel[Tally](d)
		return d
	})
	lago.RegistryAdmin.Register("p_totschool_tally.TotSchoolSession", lago.AdminPanel[TotSchoolSession]{SearchField: "Name"})
	lago.RegistryAdmin.Register("p_totschool_tally.Tally", lago.AdminPanel[Tally]{SearchField: "UserID"})
}
