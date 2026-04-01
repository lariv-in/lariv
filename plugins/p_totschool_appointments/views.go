package p_totschool_appointments

import (
	"context"
	"net/http"
	"time"

	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_users"
	"github.com/lariv-in/lago/views"
	"gorm.io/gorm"
)

// AppointmentDetailMiddleware enriches the detail view context for an appointment.
// It expects DetailView to have already loaded the concrete Appointment into the
// "appointment" context key.
func AppointmentDetailMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		rawAppt := ctx.Value("appointment")
		appointment, ok := rawAppt.(Appointment)
		if !ok {
			// If the appointment isn't present or has wrong type, fall back to the next handler.
			next.ServeHTTP(w, r)
			return
		}

		db, ok := ctx.Value("$db").(*gorm.DB)
		if !ok {
			next.ServeHTTP(w, r)
			return
		}

		if appointment.GenerationID != nil {
			ctx = context.WithValue(ctx, "GenerationPending", true)
		} else {
			ctx = context.WithValue(ctx, "GenerationPending", false)
		}

		overlapping := appointment.GetOverlappingAppointments(db)
		if len(overlapping) > 0 {
			overlapList := []map[string]any{}
			for _, o := range overlapping {
				overlapList = append(overlapList, map[string]any{
					"ID":   o.ID,
					"Name": o.Name,
					"Date": o.Datetime.Format("Jan 02, 15:04"),
				})
			}
			ctx = context.WithValue(ctx, "OverlapWarningList", overlapList)
			ctx = context.WithValue(ctx, "OverlapWarning", true)
		} else {
			// Ensure the key exists with a concrete bool so getters don't see a nil value.
			ctx = context.WithValue(ctx, "OverlapWarning", false)
		}

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func generateHandler(v *views.View) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		idStr := r.PathValue("id")
		db := r.Context().Value("$db").(*gorm.DB)
		user := r.Context().Value("$user").(p_users.User)

		var appointment Appointment
		if err := db.Where("id = ?", idStr).First(&appointment).Error; err != nil {
			http.NotFound(w, r)
			return
		}

		content, systemPrompt := buildLetterContent(db, &appointment, user.Name)
		Generate(db, appointment.ID, content, systemPrompt)

		lago.NewRedirectView("appointments.DetailRoute", map[string]getters.Getter[any]{
			"id": getters.Any(getters.Static(idStr)),
		}).ServeHTTP(w, r)
	})
}

func cancelHandler(v *views.View) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		idStr := r.PathValue("id")
		db := r.Context().Value("$db").(*gorm.DB)

		var appointment Appointment
		if err := db.Where("id = ?", idStr).First(&appointment).Error; err != nil {
			http.NotFound(w, r)
			return
		}

		if appointment.GenerationID != nil {
			CancelGeneration(db, appointment.ID)
		}

		lago.NewRedirectView("appointments.DetailRoute", map[string]getters.Getter[any]{
			"id": getters.Any(getters.Static(idStr)),
		}).ServeHTTP(w, r)
	})
}

func aiEditFormHandler(v *views.View) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		idStr := r.PathValue("id")
		db := r.Context().Value("$db").(*gorm.DB)

		var appointment Appointment
		if err := db.Where("id = ?", idStr).First(&appointment).Error; err != nil {
			http.NotFound(w, r)
			return
		}

		ctx := context.WithValue(r.Context(), "appointment", appointment)
		v.RenderPage(w, r.WithContext(ctx))
	})
}

func aiEditHandler(v *views.View) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		idStr := r.PathValue("id")
		db := r.Context().Value("$db").(*gorm.DB)

		var appointment Appointment
		if err := db.Where("id = ?", idStr).First(&appointment).Error; err != nil {
			http.NotFound(w, r)
			return
		}

		content := r.FormValue("generated_letter")
		instructions := r.FormValue("instructions")
		if content == "" || instructions == "" {
			http.Error(w, "Missing content or instructions", http.StatusBadRequest)
			return
		}

		userPrompt := "Here is the current letter content:\n\n" + content + "\n\nPlease edit this letter according to these instructions: " + instructions + "\n\nOutput only the edited text, nothing else."
		Generate(db, appointment.ID, userPrompt, letterEditorSystemPrompt)

		lago.NewRedirectView("appointments.DetailRoute", map[string]getters.Getter[any]{
			"id": getters.Any(getters.Static(idStr)),
		}).ServeHTTP(w, r)
	})
}

func FormCreatedByPatcher(v *views.View, r *http.Request, formData map[string]any) map[string]any {
	user := r.Context().Value("$user").(p_users.User)
	formData["CreatedByID"] = user.ID
	return formData
}

// scopeAppointmentsQueryToCurrentUser restricts the query to appointments created by
// the logged-in user unless they are a superuser or have the totschool_admin role.
func scopeAppointmentsQueryToCurrentUser(r *http.Request, query *gorm.DB) *gorm.DB {
	user := r.Context().Value("$user").(p_users.User)
	role, _ := r.Context().Value("$role").(string)
	if user.IsSuperuser || role == "totschool_admin" {
		return query
	}
	return query.Where("created_by_id = ?", user.ID)
}

// AppointmentListQueryPatcher applies additional filtering based on query params,
// such as the "Overlapping" checkbox in the filter form.
func AppointmentListQueryPatcher(v *views.View, r *http.Request, query *gorm.DB) *gorm.DB {
	ctx := r.Context()
	query = scopeAppointmentsQueryToCurrentUser(r, query)

	// If Overlapping=true in the parsed filter values, restrict to appointments
	// that have at least one get neighbor (same CreatedByID and within the
	// +/-30 minute window defined in GetOverlappingAppointments).
	if get, ok := ctx.Value("$get").(map[string]any); ok {
		if val, exists := get["Overlapping"]; exists {
			if b, ok := val.(bool); ok && b {
				query = WithOverlappingFilter(query)
			}
		}
		if raw, exists := get["Date"]; exists && raw != nil {
			query = applyDateFilter(raw, query)
		}
	}

	return query
}

// AppointmentTimelineQueryPatcher filters appointments for the timeline view based on
// the Date field from the timeline filter form ($get.Date). When no date is specified,
// it filters to the current calendar day (same as time.Now()).
func AppointmentTimelineQueryPatcher(v *views.View, r *http.Request, query *gorm.DB) *gorm.DB {
	ctx := r.Context()
	query = scopeAppointmentsQueryToCurrentUser(r, query)

	if get, ok := ctx.Value("$get").(map[string]any); ok {
		if raw, exists := get["Date"]; exists && raw != nil {
			switch d := raw.(type) {
			case time.Time:
				if !d.IsZero() {
					return applyDateFilter(raw, query)
				}
			case string:
				if d != "" {
					return applyDateFilter(raw, query)
				}
			}
		}
	}
	return applyDateFilter(time.Now(), query)
}

func applyDateFilter(raw any, query *gorm.DB) *gorm.DB {
	switch d := raw.(type) {
	case time.Time:
		// Filter by the same calendar day in the application's timezone.
		start := time.Date(d.Year(), d.Month(), d.Day(), 0, 0, 0, 0, d.Location())
		end := start.Add(24 * time.Hour)
		query = query.Where("datetime >= ? AND datetime < ?", start, end)
	case string:
		if d != "" {
			if parsed, err := time.Parse("2006-01-02", d); err == nil {
				start := time.Date(parsed.Year(), parsed.Month(), parsed.Day(), 0, 0, 0, 0, parsed.Location())
				end := start.Add(24 * time.Hour)
				query = query.Where("datetime >= ? AND datetime < ?", start, end)
			}
		}
	}
	return query
}

func init() {
	lago.RegistryView.Register("appointments.ListView",
		views.ListView[Appointment]("appointments")(lago.GetPageView("appointments.AppointmentTable")).
			WithQueryPatcher("appointments.list", AppointmentListQueryPatcher).
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware))

	lago.RegistryView.Register("appointments.DetailView",
		views.DetailView[Appointment]("appointment", "id")(lago.GetPageView("appointments.AppointmentDetail")).
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware).
			WithMiddleware("appointments.detail", AppointmentDetailMiddleware),
	)

	lago.RegistryView.Register("appointments.CreateView",
		views.CreateView[Appointment](lago.RoutePath("appointments.DetailRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("$id"))}))(lago.GetPageView("appointments.AppointmentCreateForm")).
			WithFormPatcher("appointments.form", FormCreatedByPatcher).
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware))

	lago.RegistryView.Register("appointments.UpdateView",
		views.DetailView[Appointment]("appointment", "id")(
			views.UpdateView[Appointment]("id", lago.RoutePath("appointments.DetailRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("$id"))}))(lago.GetPageView("appointments.AppointmentUpdateForm"))).
			WithFormPatcher("appointments.form", FormCreatedByPatcher).
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware))

	lago.RegistryView.Register("appointments.DeleteView",
		views.DetailView[Appointment]("appointment", "id")(
			views.DeleteView[Appointment]("id", lago.RoutePath("appointments.ListRoute", nil))(lago.GetPageView("appointments.AppointmentDeleteForm")).
				WithMiddleware("users.auth", p_users.AuthenticationMiddleware)))

	lago.RegistryView.Register("appointments.GenerateView",
		lago.GetPageView("appointments.AppointmentDetail").WithMethod(http.MethodPost, generateHandler).
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware))

	lago.RegistryView.Register("appointments.CancelView",
		lago.GetPageView("appointments.AppointmentDetail").WithMethod(http.MethodPost, cancelHandler).
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware))

	lago.RegistryView.Register("appointments.AiEditFormView",
		lago.GetPageView("appointments.AiEditModal").WithMethod(http.MethodGet, aiEditFormHandler).
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware))

	lago.RegistryView.Register("appointments.AiEditView",
		lago.GetPageView("appointments.AiEditModal").WithMethod(http.MethodPost, aiEditHandler).
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware))

	lago.RegistryView.Register("appointments.SelectView",
		views.ListView[Appointment]("appointments")(lago.GetPageView("appointments.AppointmentSelectionTable")).
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware))

	lago.RegistryView.Register("appointments.CardTimelineView",
		views.ListView[Appointment]("appointments")(lago.GetPageView("appointments.AppointmentCardTimeline")).
			WithQueryPatcher("appointments.timeline", AppointmentTimelineQueryPatcher).
			WithQueryPatcher("appointments.timeline_order", views.QueryPatcherOrderBy("datetime ASC")).
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware))
}
