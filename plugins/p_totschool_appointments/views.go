package p_totschool_appointments

import (
	"context"
	"fmt"
	"net/http"

	"github.com/lariv-in/getters"
	"github.com/lariv-in/lago"
	"github.com/lariv-in/p_users"
	"github.com/lariv-in/views"
	"gorm.io/gorm"
)

func redirectTo(w http.ResponseWriter, r *http.Request, url string) {
	if r.Header.Get("HX-Request") == "true" {
		w.Header().Set("HX-Redirect", url)
		w.WriteHeader(http.StatusOK)
		return
	}
	http.Redirect(w, r, url, http.StatusSeeOther)
}

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
			ctx = context.WithValue(ctx, "generation_pending", true)
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
			ctx = context.WithValue(ctx, "overlap_warning_list", overlapList)
			ctx = context.WithValue(ctx, "overlap_warning", true)
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

		content := buildLetterContent(db, &appointment, user.Name)
		Generate(db, appointment.ID, content, letterWriterSystemPrompt)

		redirectTo(w, r, fmt.Sprintf(AppUrl+"%d/", appointment.ID))
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

		redirectTo(w, r, fmt.Sprintf(AppUrl+"%d/", appointment.ID))
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

		userPrompt := fmt.Sprintf("Here is the current letter content:\n\n%s\n\nPlease edit this letter according to these instructions: %s\n\nOutput only the edited text, nothing else.", content, instructions)
		Generate(db, appointment.ID, userPrompt, letterEditorSystemPrompt)

		redirectTo(w, r, fmt.Sprintf(AppUrl+"%s/", idStr))
	})
}

func FormCreatedByPatcher(v views.View, r *http.Request, formData map[string]any) map[string]any {
	user := r.Context().Value("$user").(p_users.User)
	formData["created_by_id"] = user.ID
	fmt.Println("sldfkghl")
	return formData
}

func init() {
	lago.RegistryView.Register("appointments.ListView", p_users.AuthMiddleware(
		views.ListView[*Appointment](nil, "appointments")(lago.GetPageView("appointments.AppointmentTable"))))

	lago.RegistryView.Register("appointments.DetailView", p_users.AuthMiddleware(
		views.DetailView[*Appointment](nil, "appointment")(lago.GetPageView("appointments.AppointmentDetail").WithMethod(http.MethodGet, detailHandler))))

	lago.RegistryView.Register("appointments.CreateView", p_users.AuthMiddleware(
		views.CreateView[*Appointment](nil, AppUrl+"%v/")(lago.GetPageView("appointments.AppointmentCreateForm")).WithFormPatcher(FormCreatedByPatcher)))

	lago.RegistryView.Register("appointments.UpdateView", p_users.AuthMiddleware(
		views.UpdateView[*Appointment](nil, AppUrl+"%v/")(lago.GetPageView("appointments.AppointmentUpdateForm")).WithFormPatcher(FormCreatedByPatcher)))

	lago.RegistryView.Register("appointments.DeleteView", p_users.AuthMiddleware(
		views.DeleteView[*Appointment](nil, AppUrl)(lago.GetPageView("appointments.AppointmentDeleteForm"))))

	lago.RegistryView.Register("appointments.GenerateView", p_users.AuthMiddleware(
		lago.GetPageView("appointments.AppointmentDetail").WithMethod(http.MethodPost, generateHandler)))

	lago.RegistryView.Register("appointments.CancelView", p_users.AuthMiddleware(
		lago.GetPageView("appointments.AppointmentDetail").WithMethod(http.MethodPost, cancelHandler)))

	lago.RegistryView.Register("appointments.AiEditFormView", p_users.AuthMiddleware(
		lago.GetPageView("appointments.AiEditModal").WithMethod(http.MethodGet, aiEditFormHandler)))

	lago.RegistryView.Register("appointments.AiEditView", p_users.AuthMiddleware(
		lago.GetPageView("appointments.AiEditModal").WithMethod(http.MethodPost, aiEditHandler)))

	lago.RegistryView.Register("appointments.SelectView", p_users.AuthMiddleware(
		views.ListView[*Appointment](nil, "appointments")(lago.GetPageView("appointments.AppointmentSelectionTable"))))

	lago.RegistryView.Register("appointments.TemplateSelectView", p_users.AuthMiddleware(
		views.ListView[*LetterTemplate](nil, "templates")(lago.GetPageView("appointments.TemplateSelectionTable"))))

	lago.RegistryView.Register("appointments.CardTimelineView", p_users.AuthMiddleware(
		views.ListView[*Appointment](nil, "appointments")(lago.GetPageView("appointments.AppointmentCardTimeline"))))
}
