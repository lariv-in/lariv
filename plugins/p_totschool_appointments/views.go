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

// AppointmentDetailCtxMiddleware enriches detail context after MiddlewareDetail loads "appointment".
type AppointmentDetailCtxMiddleware struct{}

func (AppointmentDetailCtxMiddleware) Next(_ views.View, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		rawAppt := ctx.Value("appointment")
		appointment, ok := rawAppt.(Appointment)
		if !ok {
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
					"Date": o.Datetime,
				})
			}
			ctx = context.WithValue(ctx, "OverlapWarningList", overlapList)
			ctx = context.WithValue(ctx, "OverlapWarning", true)
		} else {
			ctx = context.WithValue(ctx, "OverlapWarning", false)
		}

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func redirectAppointmentDetail(w http.ResponseWriter, r *http.Request, idStr string) bool {
	url, err := getters.IfOr(lago.RoutePath("appointments.DetailRoute", map[string]getters.Getter[any]{
		"id": getters.Any(getters.Static(idStr)),
	}), r.Context(), "")
	if err != nil || url == "" {
		http.NotFound(w, r)
		return false
	}
	lago.Redirect(w, r, url)
	return true
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

		appointment, err := gorm.G[Appointment](db).Where("id = ?", idStr).First(r.Context())
		if err != nil {
			http.NotFound(w, r)
			return
		}

		content, systemPrompt := buildLetterContent(db, r.Context(), &appointment, user.Name)
		Generate(db, appointment.ID, content, systemPrompt)

		redirectAppointmentDetail(w, r, idStr)
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

		appointment, err := gorm.G[Appointment](db).Where("id = ?", idStr).First(r.Context())
		if err != nil {
			http.NotFound(w, r)
			return
		}

		if appointment.GenerationID != nil {
			CancelGeneration(db, appointment.ID)
		}

		redirectAppointmentDetail(w, r, idStr)
	})
}

func aiEditFormHandler(v *views.View) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		idStr := r.PathValue("id")
		db := r.Context().Value("$db").(*gorm.DB)

		appointment, err := gorm.G[Appointment](db).Where("id = ?", idStr).First(r.Context())
		if err != nil {
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

		appointment, err := gorm.G[Appointment](db).Where("id = ?", idStr).First(r.Context())
		if err != nil {
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

		redirectAppointmentDetail(w, r, idStr)
	})
}

type appointmentFormCreatedByPatcher struct{}

func (appointmentFormCreatedByPatcher) Patch(_ views.View, r *http.Request, formData map[string]any, formErrors map[string]error) (map[string]any, map[string]error) {
	user := r.Context().Value("$user").(p_users.User)
	formData["CreatedByID"] = user.ID
	return formData, formErrors
}

func scopeAppointmentsQueryToCurrentUser(r *http.Request, query gorm.ChainInterface[Appointment]) gorm.ChainInterface[Appointment] {
	user := r.Context().Value("$user").(p_users.User)
	role, _ := r.Context().Value("$role").(string)
	if user.IsSuperuser || role == "totschool_admin" {
		return query
	}
	return query.Where("created_by_id = ?", user.ID)
}

type appointmentListQueryPatcher struct{}

func (appointmentListQueryPatcher) Patch(_ views.View, r *http.Request, query gorm.ChainInterface[Appointment]) gorm.ChainInterface[Appointment] {
	ctx := r.Context()
	query = scopeAppointmentsQueryToCurrentUser(r, query)

	if get, ok := ctx.Value("$get").(map[string]any); ok {
		if val, exists := get["Overlapping"]; exists {
			if b, ok := val.(bool); ok && b {
				query = WithOverlappingFilterChain(query)
			}
		}
		if raw, exists := get["Date"]; exists && raw != nil {
			query = applyDateFilterChain(raw, query)
		}
	}

	return query
}

type appointmentTimelineQueryPatcher struct{}

func (appointmentTimelineQueryPatcher) Patch(_ views.View, r *http.Request, query gorm.ChainInterface[Appointment]) gorm.ChainInterface[Appointment] {
	ctx := r.Context()
	query = scopeAppointmentsQueryToCurrentUser(r, query)

	if get, ok := ctx.Value("$get").(map[string]any); ok {
		if raw, exists := get["Date"]; exists && raw != nil {
			switch d := raw.(type) {
			case time.Time:
				if !d.IsZero() {
					return applyDateFilterChain(raw, query)
				}
			case string:
				if d != "" {
					return applyDateFilterChain(raw, query)
				}
			}
		}
	}
	return applyDateFilterChain(time.Now(), query)
}

func applyDateFilterChain(raw any, query gorm.ChainInterface[Appointment]) gorm.ChainInterface[Appointment] {
	switch d := raw.(type) {
	case time.Time:
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
		lago.GetPageView("appointments.AppointmentTable").
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware{}).
			WithMiddleware("appointments.list", views.MiddlewareList[Appointment]{
				Key: getters.Static("appointments"),
				QueryPatchers: views.QueryPatchers[Appointment]{
					{Key: "appointments.list", Value: appointmentListQueryPatcher{}},
				},
			}))

	lago.RegistryView.Register("appointments.DetailView",
		lago.GetPageView("appointments.AppointmentDetail").
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware{}).
			WithMiddleware("appointments.detail", views.MiddlewareDetail[Appointment]{
				Key:          getters.Static("appointment"),
				PathParamKey: getters.Static("id"),
			}).
			WithMiddleware("appointments.detail_ctx", AppointmentDetailCtxMiddleware{}),
	)

	lago.RegistryView.Register("appointments.CreateView",
		lago.GetPageView("appointments.AppointmentCreateForm").
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware{}).
			WithMiddleware("appointments.create", views.MiddlewareCreate[Appointment]{
				SuccessURL: lago.RoutePath("appointments.DetailRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("$id")),
				}),
				FormPatchers: views.FormPatchers{
					{Key: "appointments.form", Value: appointmentFormCreatedByPatcher{}},
				},
			}))

	lago.RegistryView.Register("appointments.UpdateView",
		lago.GetPageView("appointments.AppointmentUpdateForm").
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware{}).
			WithMiddleware("appointments.detail", views.MiddlewareDetail[Appointment]{
				Key:          getters.Static("appointment"),
				PathParamKey: getters.Static("id"),
			}).
			WithMiddleware("appointments.update", views.MiddlewareUpdate[Appointment]{
				Key: getters.Static("appointment"),
				SuccessURL: lago.RoutePath("appointments.DetailRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("appointment.ID")),
				}),
				FormPatchers: views.FormPatchers{
					{Key: "appointments.form", Value: appointmentFormCreatedByPatcher{}},
				},
			}))

	lago.RegistryView.Register("appointments.DeleteView",
		lago.GetPageView("appointments.AppointmentDeleteForm").
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware{}).
			WithMiddleware("appointments.detail", views.MiddlewareDetail[Appointment]{
				Key:          getters.Static("appointment"),
				PathParamKey: getters.Static("id"),
			}).
			WithMiddleware("appointments.delete", views.MiddlewareDelete[Appointment]{
				Key:        getters.Static("appointment"),
				SuccessURL: lago.RoutePath("appointments.ListRoute", nil),
			}))

	lago.RegistryView.Register("appointments.GenerateView",
		lago.GetPageView("appointments.AppointmentDetail").
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware{}).
			WithMiddleware("appointments.generate", views.MethodMiddleware{
				Method:  http.MethodPost,
				Handler: generateHandler,
			}))

	lago.RegistryView.Register("appointments.CancelView",
		lago.GetPageView("appointments.AppointmentDetail").
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware{}).
			WithMiddleware("appointments.cancel", views.MethodMiddleware{
				Method:  http.MethodPost,
				Handler: cancelHandler,
			}))

	lago.RegistryView.Register("appointments.AiEditFormView",
		lago.GetPageView("appointments.AiEditModal").
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware{}).
			WithMiddleware("appointments.ai_edit_form", views.MethodMiddleware{
				Method:  http.MethodGet,
				Handler: aiEditFormHandler,
			}))

	lago.RegistryView.Register("appointments.AiEditView",
		lago.GetPageView("appointments.AiEditModal").
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware{}).
			WithMiddleware("appointments.ai_edit", views.MethodMiddleware{
				Method:  http.MethodPost,
				Handler: aiEditHandler,
			}))

	lago.RegistryView.Register("appointments.SelectView",
		lago.GetPageView("appointments.AppointmentSelectionTable").
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware{}).
			WithMiddleware("appointments.select_list", views.MiddlewareList[Appointment]{
				Key: getters.Static("appointments"),
			}))

	lago.RegistryView.Register("appointments.CardTimelineView",
		lago.GetPageView("appointments.AppointmentCardTimeline").
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware{}).
			WithMiddleware("appointments.timeline", views.MiddlewareList[Appointment]{
				Key: getters.Static("appointments"),
				QueryPatchers: views.QueryPatchers[Appointment]{
					{Key: "appointments.timeline", Value: appointmentTimelineQueryPatcher{}},
					{Key: "appointments.timeline_order", Value: views.QueryPatcherOrderBy[Appointment]{Order: "datetime ASC"}},
				},
			}))
}
