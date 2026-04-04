package p_totschool_tally

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_users"
	"github.com/lariv-in/lago/views"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// getSessionFromEnvironment looks up the session selected in the $environment cookie
// (session id), falling back to the current quarter if none is selected.
func getSessionFromEnvironment(db *gorm.DB, ctx context.Context) TotSchoolSession {
	if envMap, ok := ctx.Value("$environment").(map[string]string); ok {
		if raw, exists := envMap["session"]; exists && raw != "" {
			if id, err := strconv.ParseUint(raw, 10, 64); err == nil && id > 0 {
				session, qerr := gorm.G[TotSchoolSession](db).Where("id = ?", uint(id)).First(ctx)
				if qerr == nil {
					return session
				} else {
					slog.Error("getSessionFromEnvironment: failed to load session by id from $environment",
						"id", raw,
						"error", qerr,
					)
				}
			} else {
				// Legacy: cookie held session name.
				session, err := gorm.G[TotSchoolSession](db).Where("name = ?", raw).First(ctx)
				if err == nil {
					return session
				}
			}
		}
	}
	slog.Error("getSessionFromEnvironment: no session found in $environment", "environment", ctx.Value("$environment"))
	return EnsureSessionForDate(db, time.Now())
}

// TallyDashboardHandler displays user stats.
func TallyDashboardHandler(v *views.View) http.Handler {
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

		session := getSessionFromEnvironment(db, r.Context())

		dashboard := GetDashboardStats(db, userID, &session)

		data := map[string]any{
			"Dashboard": dashboard,
			"Session":   session,
		}

		// For non-admin users, provide WhatsApp report data for the dashboard.
		if !user.IsSuperuser && roleName != "totschool_admin" {
			data["WhatsappReport"] = GetWhatsappReportData(db, user.ID)
		}

		ctx := context.WithValue(r.Context(), "$in", data)

		v.RenderPage(w, r.WithContext(ctx))
	})
}

// TallyLeaderboardHandler displays top users per metric.
func TallyLeaderboardHandler(v *views.View) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := r.Context().Value("$user").(p_users.User)
		db := r.Context().Value("$db").(*gorm.DB)

		session := getSessionFromEnvironment(db, r.Context())

		leaderboards := GetLeaderboards(db, &user.ID, &session)

		ctx := context.WithValue(r.Context(), "$in", map[string]any{
			"Leaderboards": leaderboards,
			"Title":        fmt.Sprintf("Leaderboard for %s", session.Name),
		})

		v.RenderPage(w, r.WithContext(ctx))
	})
}

// TallyDetailQueryPatcher scopes the detail query so that non-admin users
// can only see their own tallies, while admins/superusers can see all.
type tallyDetailQueryPatcher struct{}

func (tallyDetailQueryPatcher) Patch(_ views.View, r *http.Request, query gorm.ChainInterface[Tally]) gorm.ChainInterface[Tally] {
	user, ok := r.Context().Value("$user").(p_users.User)
	if !ok {
		return query
	}

	db := r.Context().Value("$db").(*gorm.DB)
	var roleName string
	if !user.IsSuperuser {
		db.Model(&p_users.Role{}).Where("id = ?", user.RoleID).Select("name").Scan(&roleName)
	} else {
		roleName = "superuser"
	}

	if !user.IsSuperuser && roleName != "totschool_admin" {
		query = query.Where("user_id = ?", user.ID)
	}
	return query
}

var TallyDetailQueryPatcher views.QueryPatcher[Tally] = tallyDetailQueryPatcher{}

type requireAdminLayer struct{}

func (requireAdminLayer) Next(_ views.View, next http.Handler) http.Handler {
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

// TallyListQueryPatcher scopes and filters the tallies list for the generic ListView.
type tallyListQueryPatcher struct{}

func (tallyListQueryPatcher) Patch(_ views.View, r *http.Request, query gorm.ChainInterface[Tally]) gorm.ChainInterface[Tally] {
	ctx := r.Context()

	rawUser := ctx.Value("$user")
	if rawUser == nil {
		slog.Error("TallyListQueryPatcher: missing $user in context – auth layer not applied?")
		panic("TallyListQueryPatcher: $user is nil in context")
	}
	user, ok := rawUser.(p_users.User)
	if !ok {
		slog.Error("TallyListQueryPatcher: $user has unexpected type",
			"type", fmt.Sprintf("%T", rawUser),
		)
		panic("TallyListQueryPatcher: $user has wrong type in context")
	}

	dbVal := ctx.Value("$db")
	db, ok := dbVal.(*gorm.DB)
	if !ok || db == nil {
		slog.Error("TallyListQueryPatcher: missing or invalid $db in context",
			"type", fmt.Sprintf("%T", dbVal),
		)
		panic("TallyListQueryPatcher: $db is nil or wrong type in context")
	}
	// Always join the related user so table columns can access User.Name.
	query = query.Joins(clause.JoinTarget{Association: "User"}, nil)

	// Restrict to the current session.
	session := getSessionFromEnvironment(db, ctx)
	query = query.Where("date >= ? AND date <= ?", session.Start, session.End)

	// Role-based scoping: non-admin users can only see their own tallies.
	var roleName string
	if !user.IsSuperuser {
		db.Model(&p_users.Role{}).Where("id = ?", user.RoleID).Select("name").Scan(&roleName)
	}
	isAdmin := user.IsSuperuser || roleName == "totschool_admin"
	if !isAdmin {
		query = query.Where("user_id = ?", user.ID)
	}

	// Apply filters from $get (populated by the generic ListView from query params + filter form).
	if getMap, ok := ctx.Value("$get").(map[string]any); ok {
		// User filter: only effective for admins/superusers.
		if isAdmin {
			if val, ok := getMap["UserID"]; ok && val != nil {
				switch v := val.(type) {
				case uint:
					if v != 0 {
						query = query.Where("user_id = ?", v)
					}
				case string:
					if v != "" {
						if parsed, err := strconv.ParseUint(v, 10, 64); err == nil && parsed != 0 {
							query = query.Where("user_id = ?", uint(parsed))
						}
					}
				}
			}
		}

		// Date filter: when provided, narrow to that specific calendar day.
		if raw, ok := getMap["Date"]; ok && raw != nil {
			switch d := raw.(type) {
			case time.Time:
				if !d.IsZero() {
					start := time.Date(d.Year(), d.Month(), d.Day(), 0, 0, 0, 0, d.Location())
					end := start.Add(24 * time.Hour)
					query = query.Where("date >= ? AND date < ?", start, end)
				}
			case string:
				if d != "" {
					if parsed, err := time.Parse("2006-01-02", d); err == nil {
						start := time.Date(parsed.Year(), parsed.Month(), parsed.Day(), 0, 0, 0, 0, parsed.Location())
						end := start.Add(24 * time.Hour)
						query = query.Where("date >= ? AND date < ?", start, end)
					}
				}
			}
		}
	}

	return query
}

var TallyListQueryPatcher views.QueryPatcher[Tally] = tallyListQueryPatcher{}

// TallyDailyFormHandler handles form submission for the logged-in user's daily tally.
func TallyDailyFormHandler(v *views.View) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := r.Context().Value("$user").(p_users.User)
		db := r.Context().Value("$db").(*gorm.DB)
		today := time.Now().Truncate(24 * time.Hour)

		tally, err := gorm.G[Tally](db).Where("user_id = ? AND date = ?", user.ID, today).First(r.Context())
		if err != nil {
			tally = Tally{UserID: user.ID, Date: today}
		}

		if r.Method == http.MethodGet {
			// Pre-fill the form by projecting the loaded tally into $in so
			// InputNumber fields using GetterKey("$in.*") can resolve values.
			ctx := context.WithValue(r.Context(), getters.ContextKeyIn, tally)
			v.RenderPage(w, r.WithContext(ctx))
			return
		}

		values, fieldErrors, err := v.ParseForm(w, r)
		if err != nil {
			return
		}

		if len(fieldErrors) != 0 {
			ctx := context.WithValue(r.Context(), "$in", map[string]any{"Tally": tally})
			ctx = views.ContextWithErrorsAndValues(ctx, values, fieldErrors)
			v.RenderPage(w, r.WithContext(ctx))
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
			_ = gorm.G[Tally](db).Create(r.Context(), &tally)
		} else {
			db.Save(&tally)
		}

		views.HtmxRedirect(w, r, "/tally/", http.StatusSeeOther)
	})
}

func init() {
	lago.RegistryView.Register("tally.TallyDashboardView",
		lago.GetPageView("tally.TallyDashboard").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("tally.dashboard", views.MethodLayer{
				Method:  http.MethodGet,
				Handler: TallyDashboardHandler,
			}))

	lago.RegistryView.Register("tally.TallyLeaderboardView",
		lago.GetPageView("tally.TallyLeaderboard").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("tally.leaderboard", views.MethodLayer{
				Method:  http.MethodGet,
				Handler: TallyLeaderboardHandler,
			}))

	lago.RegistryView.Register("tally.TallyListView",
		lago.GetPageView("tally.TallyTable").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("tally.list", views.LayerList[Tally]{
				Key: getters.Static("Tallies"),
				QueryPatchers: views.QueryPatchers[Tally]{
					{Key: "tally.list", Value: TallyListQueryPatcher},
				},
			}))

	lago.RegistryView.Register("tally.TallyDailyFormView",
		lago.GetPageView("tally.TallyDailyForm").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("tally.daily_form_get", views.MethodLayer{
				Method:  http.MethodGet,
				Handler: TallyDailyFormHandler,
			}).
			WithLayer("tally.daily_form_post", views.MethodLayer{
				Method:  http.MethodPost,
				Handler: TallyDailyFormHandler,
			}))

	lago.RegistryView.Register("tally.TallyCreateView",
		lago.GetPageView("tally.TallyCreateForm").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("tally.admin", requireAdminLayer{}).
			WithLayer("tally.create", views.LayerCreate[Tally]{
				SuccessURL: lago.RoutePath("tally.TallyListRoute", nil),
			}))

	lago.RegistryView.Register("tally.TallyUpdateView",
		lago.GetPageView("tally.TallyUpdateForm").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("tally.admin", requireAdminLayer{}).
			WithLayer("tally.detail", views.LayerDetail[Tally]{
				Key:          getters.Static("Tally"),
				PathParamKey: getters.Static("id"),
			}).
			WithLayer("tally.update", views.LayerUpdate[Tally]{
				Key:        getters.Static("Tally"),
				SuccessURL: lago.RoutePath("tally.TallyListRoute", nil),
			}))

	lago.RegistryView.Register("tally.TallyDeleteView",
		lago.GetPageView("tally.TallyDeleteForm").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("tally.admin", requireAdminLayer{}).
			WithLayer("tally.detail", views.LayerDetail[Tally]{
				Key:          getters.Static("Tally"),
				PathParamKey: getters.Static("id"),
			}).
			WithLayer("tally.delete", views.LayerDelete[Tally]{
				Key:        getters.Static("Tally"),
				SuccessURL: lago.RoutePath("tally.TallyListRoute", nil),
			}))

	lago.RegistryView.Register("tally.TallyDetailView",
		lago.GetPageView("tally.TallyDetail").
			WithLayer("users.auth", p_users.AuthenticationLayer{}).
			WithLayer("tally.detail", views.LayerDetail[Tally]{
				Key:          getters.Static("Tally"),
				PathParamKey: getters.Static("id"),
				QueryPatchers: views.QueryPatchers[Tally]{
					{Key: "tally.detail", Value: TallyDetailQueryPatcher},
				},
			}))
}
