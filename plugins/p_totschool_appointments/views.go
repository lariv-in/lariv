package p_totschool_appointments

import (
	"context"
	"net/http"

	"github.com/lariv-in/getters"
	"github.com/lariv-in/lago"
	"github.com/lariv-in/p_users"
	"github.com/lariv-in/views"
	"gorm.io/gorm"
)

func detailHandler(v views.View) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		idStr := r.PathValue("id")
		db := r.Context().Value("$db").(*gorm.DB)

		var appointment Appointment
		// Verify created_by manually if not superuser or admin
		err := db.Where("id = ?", idStr).First(&appointment).Error
		if err != nil {
			http.NotFound(w, r)
			return
		}

		ctx := r.Context()

		if appointment.GenerationID != nil {
			ctx = context.WithValue(ctx, "GenerationPending", true)
		}

		appointmentMap := getters.MapFromStruct(&appointment)

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
		}

		ctx = context.WithValue(ctx, "appointment", appointmentMap)
		v.RenderPage(w, r.WithContext(ctx))
	})
}

func generateHandler(v views.View) http.Handler {
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

		lago.NewRedirectView("appointments.DetailRoute", map[string]getters.Getter{
			"id": getters.GetterStatic(idStr),
		}).ServeHTTP(w, r)
	})
}

func cancelHandler(v views.View) http.Handler {
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

		lago.NewRedirectView("appointments.DetailRoute", map[string]getters.Getter{
			"id": getters.GetterStatic(idStr),
		}).ServeHTTP(w, r)
	})
}

func aiEditFormHandler(v views.View) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		idStr := r.PathValue("id")
		db := r.Context().Value("$db").(*gorm.DB)

		var appointment Appointment
		if err := db.Where("id = ?", idStr).First(&appointment).Error; err != nil {
			http.NotFound(w, r)
			return
		}

		ctx := context.WithValue(r.Context(), "appointment", getters.MapFromStruct(&appointment))
		v.RenderPage(w, r.WithContext(ctx))
	})
}

func aiEditHandler(v views.View) http.Handler {
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

		lago.NewRedirectView("appointments.DetailRoute", map[string]getters.Getter{
			"id": getters.GetterStatic(idStr),
		}).ServeHTTP(w, r)
	})
}

func FormCreatedByPatcher(v views.View, r *http.Request, formData map[string]any) map[string]any {
	user := r.Context().Value("$user").(p_users.User)
	formData["CreatedByID"] = user.ID
	return formData
}

func init() {
	lago.RegistryView.Register("appointments.ListView", p_users.AuthMiddleware(
		views.ListView[Appointment]("appointments")(lago.GetPageView("appointments.AppointmentTable"))))

	lago.RegistryView.Register("appointments.DetailView", p_users.AuthMiddleware(
		views.DetailView[Appointment]("appointment")(lago.GetPageView("appointments.AppointmentDetail").WithMethod(http.MethodGet, detailHandler))))

	lago.RegistryView.Register("appointments.CreateView", p_users.AuthMiddleware(
		views.CreateView[Appointment](lago.GetterRoutePath("appointments.DetailRoute", map[string]getters.Getter{"id": getters.GetterKey("$id")}))(lago.GetPageView("appointments.AppointmentCreateForm")).WithFormPatcher(FormCreatedByPatcher)))

	lago.RegistryView.Register("appointments.UpdateView", p_users.AuthMiddleware(
		views.UpdateView[Appointment](lago.GetterRoutePath("appointments.DetailRoute", map[string]getters.Getter{"id": getters.GetterKey("$id")}))(lago.GetPageView("appointments.AppointmentUpdateForm")).WithFormPatcher(FormCreatedByPatcher)))

	lago.RegistryView.Register("appointments.DeleteView", p_users.AuthMiddleware(
		views.DeleteView[Appointment](lago.GetterRoutePath("appointments.ListRoute", nil))(lago.GetPageView("appointments.AppointmentDeleteForm"))))

	lago.RegistryView.Register("appointments.GenerateView", p_users.AuthMiddleware(
		lago.GetPageView("appointments.AppointmentDetail").WithMethod(http.MethodPost, generateHandler)))

	lago.RegistryView.Register("appointments.CancelView", p_users.AuthMiddleware(
		lago.GetPageView("appointments.AppointmentDetail").WithMethod(http.MethodPost, cancelHandler)))

	lago.RegistryView.Register("appointments.AiEditFormView", p_users.AuthMiddleware(
		lago.GetPageView("appointments.AiEditModal").WithMethod(http.MethodGet, aiEditFormHandler)))

	lago.RegistryView.Register("appointments.AiEditView", p_users.AuthMiddleware(
		lago.GetPageView("appointments.AiEditModal").WithMethod(http.MethodPost, aiEditHandler)))

	lago.RegistryView.Register("appointments.SelectView", p_users.AuthMiddleware(
		views.ListView[Appointment]("appointments")(lago.GetPageView("appointments.AppointmentSelectionTable"))))

	lago.RegistryView.Register("appointments.CardTimelineView", p_users.AuthMiddleware(
		views.ListView[Appointment]("appointments")(lago.GetPageView("appointments.AppointmentCardTimeline"))))
}
