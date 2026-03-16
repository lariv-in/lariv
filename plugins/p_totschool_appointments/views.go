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

func detailHandler(v *views.View) http.Handler {
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

		// Store the concrete Appointment in context; Detail[Appointment] and
		// components.FormComponent[Appointment] use GetterKey[Appointment]("appointment")
		// and will map it into $in themselves.
		ctx = context.WithValue(ctx, "appointment", appointment)
		v.RenderPage(w, r.WithContext(ctx))
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
			"id": getters.GetterAny(getters.GetterStatic(idStr)),
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
			"id": getters.GetterAny(getters.GetterStatic(idStr)),
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
			"id": getters.GetterAny(getters.GetterStatic(idStr)),
		}).ServeHTTP(w, r)
	})
}

func FormCreatedByPatcher(v *views.View, r *http.Request, formData map[string]any) map[string]any {
	user := r.Context().Value("$user").(p_users.User)
	formData["CreatedByID"] = user.ID
	return formData
}

func init() {
	lago.RegistryView.Register("appointments.ListView",
		views.ListView[Appointment]("appointments")(lago.GetPageView("appointments.AppointmentTable")).
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware))

	lago.RegistryView.Register("appointments.DetailView",
		views.DetailView[Appointment]("appointment")(lago.GetPageView("appointments.AppointmentDetail").WithMethod(http.MethodGet, detailHandler)).
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware))

	lago.RegistryView.Register("appointments.CreateView",
		views.CreateView[Appointment](lago.GetterRoutePath("appointments.DetailRoute", map[string]getters.Getter[any]{"id": getters.GetterAny(getters.GetterKey[string]("$id"))}))(lago.GetPageView("appointments.AppointmentCreateForm")).
			WithFormPatcher(FormCreatedByPatcher).
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware))

	lago.RegistryView.Register("appointments.UpdateView",
		views.UpdateView[Appointment](lago.GetterRoutePath("appointments.DetailRoute", map[string]getters.Getter[any]{"id": getters.GetterAny(getters.GetterKey[string]("$id"))}))(lago.GetPageView("appointments.AppointmentUpdateForm")).
			WithFormPatcher(FormCreatedByPatcher).
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware))

	lago.RegistryView.Register("appointments.DeleteView",
		views.DeleteView[Appointment](lago.GetterRoutePath("appointments.ListRoute", nil))(lago.GetPageView("appointments.AppointmentDeleteForm")).
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware))

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
			WithMiddleware("users.auth", p_users.AuthenticationMiddleware))
}
