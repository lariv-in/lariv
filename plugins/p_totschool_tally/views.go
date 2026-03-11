package p_totschool_tally

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/lariv-in/getters"
	"github.com/lariv-in/lago"
	"github.com/lariv-in/p_users"
	"github.com/lariv-in/views"
	"gorm.io/gorm"
)

// TallyDashboardHandler displays user stats.
func TallyDashboardHandler(v views.View) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := r.Context().Value("$user").(p_users.User)
		db := r.Context().Value("$db").(*gorm.DB)

		var roleName string
		if !user.IsSuperuser {
			db.Model(&p_users.Role{}).Where("id = ?", user.RoleID).Select("name").Scan(&roleName)
		}

		var userID *uint
		if !user.IsSuperuser && roleName != "totschool_admin" {
			userID = &user.ID
		}

		session := EnsureSessionForDate(db, time.Now())

		dashboard := GetDashboardStats(db, userID, &session)

		ctx := context.WithValue(r.Context(), "$in", map[string]any{
			"dashboard": dashboard,
			"session":   session,
		})

		v.RenderPage(w, r.WithContext(ctx))
	})
}

// TallyLeaderboardHandler displays top users per metric.
func TallyLeaderboardHandler(v views.View) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := r.Context().Value("$user").(p_users.User)
		db := r.Context().Value("$db").(*gorm.DB)

		session := EnsureSessionForDate(db, time.Now())

		leaderboards := GetLeaderboards(db, &user.ID, &session)

		ctx := context.WithValue(r.Context(), "$in", map[string]any{
			"leaderboards": leaderboards,
			"title":        fmt.Sprintf("Leaderboard for %s", session.Name),
		})

		v.RenderPage(w, r.WithContext(ctx))
	})
}

// Ensure the user owns or is admin for the queried tally
func getTallyOr404(w http.ResponseWriter, r *http.Request, db *gorm.DB, tallyID string, user p_users.User) *Tally {
	var roleName string
	if !user.IsSuperuser {
		db.Model(&p_users.Role{}).Where("id = ?", user.RoleID).Select("name").Scan(&roleName)
	} else {
		roleName = "superuser"
	}

	var tally Tally
	query := db.Where("id = ?", tallyID)
	if !user.IsSuperuser && roleName != "totschool_admin" {
		query = query.Where("user_id = ?", user.ID)
	}

	if err := query.First(&tally).Error; err != nil {
		http.Error(w, "Tally not found", http.StatusNotFound)
		return nil
	}
	return &tally
}

// RequireAdmin middleware to restrict access to admins only.
func RequireAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, ok := r.Context().Value("$user").(p_users.User)
		if !ok {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		db := r.Context().Value("$db").(*gorm.DB)
		var roleName string
		if !user.IsSuperuser {
			db.Model(&p_users.Role{}).Where("id = ?", user.RoleID).Select("name").Scan(&roleName)
		}

		if !user.IsSuperuser && roleName != "totschool_admin" {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// TallyListHandler lists all tallies
func TallyListHandler(v views.View) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := r.Context().Value("$user").(p_users.User)
		db := r.Context().Value("$db").(*gorm.DB)

		session := EnsureSessionForDate(db, time.Now())

		var roleName string
		if !user.IsSuperuser {
			db.Model(&p_users.Role{}).Where("id = ?", user.RoleID).Select("name").Scan(&roleName)
		}

		query := db.Model(&Tally{}).Joins("User").Where("date >= ? AND date <= ?", session.Start, session.End)
		if !user.IsSuperuser && roleName != "totschool_admin" {
			query = query.Where("user_id = ?", user.ID)
		}

		var tallies []Tally
		query.Order("date DESC").Find(&tallies)

		ctx := context.WithValue(r.Context(), "$in", map[string]any{
			"tallies": tallies,
		})
		v.RenderPage(w, r.WithContext(ctx))
	})
}

// TallyDailyFormHandler handles form submission for the logged-in user's daily tally.
func TallyDailyFormHandler(v views.View) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := r.Context().Value("$user").(p_users.User)
		db := r.Context().Value("$db").(*gorm.DB)
		today := time.Now().Truncate(24 * time.Hour)

		var tally Tally
		err := db.Where("user_id = ? AND date = ?", user.ID, today).First(&tally).Error
		if err != nil {
			tally = Tally{UserID: user.ID, Date: today}
		}

		if r.Method == http.MethodGet {
			ctx := context.WithValue(r.Context(), "$in", map[string]any{"tally": tally})
			v.RenderPage(w, r.WithContext(ctx))
			return
		}

		values, fieldErrors, failed := v.ParseForm(w, r)
		if failed {
			return
		}

		if views.HasErrors(fieldErrors) {
			ctx := context.WithValue(r.Context(), "$in", map[string]any{"tally": tally})
			v.RenderWithErrors(w, r.WithContext(ctx), fieldErrors, values)
			return
		}

		// Update fields using GORM Updates instead of mass assignment map to respect integer types directly
		tally.Visits, _ = values["Visits"].(int)
		tally.Appointments, _ = values["Appointments"].(int)
		tally.Leads, _ = values["Leads"].(int)
		tally.Presentations, _ = values["Presentations"].(int)
		tally.Demos, _ = values["Demos"].(int)
		tally.Letters, _ = values["Letters"].(int)
		tally.FollowUps, _ = values["FollowUps"].(int)
		tally.Proposals, _ = values["Proposals"].(int)
		tally.Policies, _ = values["Policies"].(int)
		tally.Premium, _ = values["Premium"].(int)

		if tally.ID == 0 {
			db.Create(&tally)
		} else {
			db.Save(&tally)
		}

		http.Redirect(w, r, "/tally/", http.StatusSeeOther)
	})
}

func init() {
	lago.RegistryView.Register("tally.TallyDashboardView", p_users.AuthMiddleware(lago.GetPageView("tally.TallyDashboard").WithMethod(http.MethodGet, TallyDashboardHandler)))
	lago.RegistryView.Register("tally.TallyLeaderboardView", p_users.AuthMiddleware(lago.GetPageView("tally.TallyLeaderboard").WithMethod(http.MethodGet, TallyLeaderboardHandler)))
	lago.RegistryView.Register("tally.TallyListView", p_users.AuthMiddleware(lago.GetPageView("tally.TallyTable").WithMethod(http.MethodGet, TallyListHandler)))
	lago.RegistryView.Register("tally.TallyDailyFormView", p_users.AuthMiddleware(lago.GetPageView("tally.TallyDailyForm").WithMethod(http.MethodGet, TallyDailyFormHandler).WithMethod(http.MethodPost, TallyDailyFormHandler)))

	// Admin CRUD mappings using standard views
	lago.RegistryView.Register("tally.TallyCreateView", p_users.AuthMiddleware(RequireAdmin(views.CreateView(Tally{}, "/tally/")(lago.GetPageView("tally.TallyCreateForm")))))
	lago.RegistryView.Register("tally.TallyUpdateView", p_users.AuthMiddleware(RequireAdmin(views.UpdateView(Tally{}, "/tally/")(lago.GetPageView("tally.TallyUpdateForm")))))
	lago.RegistryView.Register("tally.TallyDeleteView", p_users.AuthMiddleware(RequireAdmin(views.DeleteView(Tally{}, "/tally/")(lago.GetPageView("tally.TallyDeleteForm")))))

	// Detail View allows access if user owns it or is admin
	lago.RegistryView.Register("tally.TallyDetailView", p_users.AuthMiddleware(func(v views.View) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			idStr := r.PathValue("id")
			db := r.Context().Value("$db").(*gorm.DB)
			user := r.Context().Value("$user").(p_users.User)

			tally := getTallyOr404(w, r, db, idStr, user)
			if tally == nil {
				return
			}
			ctx := context.WithValue(r.Context(), "$in", map[string]any{"tally": getters.MapFromStruct(tally)})
			v.RenderPage(w, r.WithContext(ctx))
		})
	}(lago.GetPageView("tally.TallyDetail")))) // Using new TallyDetail component
}
