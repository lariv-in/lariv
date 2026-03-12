package p_totschool_appointments

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/lariv-in/components"
	"github.com/lariv-in/getters"
	"github.com/lariv-in/lago"
	"github.com/lariv-in/p_users"
	"github.com/lariv-in/views"
	"gorm.io/gorm"
)

func applyFormValues(a *Appointment, values map[string]any) {
	if val, ok := values["name"].(string); ok {
		a.Name = val
	}
	if val, ok := values["location"].(string); ok {
		a.Location = val
	}
	if val, ok := values["phone"].(string); ok {
		a.Phone = val
	}
	if val, ok := values["remarks"].(string); ok {
		a.Remarks = val
	}
	if val, ok := values["extra_info"].(string); ok {
		a.ExtraInfo = val
	}
	if val, ok := values["datetime"].(time.Time); ok {
		a.Datetime = val
	}
}

func appointmentScope(db *gorm.DB, user p_users.User) *gorm.DB {
	if user.IsSuperuser {
		return db
	}
	var roleName string
	db.Model(&p_users.Role{}).Where("id = ?", user.RoleID).Select("name").Scan(&roleName)
	if roleName == "totschool_admin" {
		return db
	}
	return db.Where("created_by_id = ?", user.ID)
}

func getAppointmentOr404(w http.ResponseWriter, r *http.Request, db *gorm.DB, idStr string, user p_users.User) *Appointment {
	var a Appointment
	err := appointmentScope(db, user).Where("id = ?", idStr).First(&a).Error
	if err != nil {
		http.NotFound(w, r)
		return nil
	}
	return &a
}

func redirectTo(w http.ResponseWriter, r *http.Request, url string) {
	if r.Header.Get("HX-Request") == "true" {
		w.Header().Set("HX-Redirect", url)
		w.WriteHeader(http.StatusOK)
		return
	}
	http.Redirect(w, r, url, http.StatusSeeOther)
}

func listHandler(v views.View) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		db := r.Context().Value("$db").(*gorm.DB)
		user := r.Context().Value("$user").(p_users.User)

		query := appointmentScope(db, user).Model(&Appointment{})

		pageStr := r.URL.Query().Get("page")
		pageNum := 1
		if pageStr != "" {
			if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
				pageNum = p
			}
		}
		pageSize := 12

		if name := r.URL.Query().Get("name"); name != "" {
			query = query.Where("name LIKE ?", "%"+name+"%")
		}
		if location := r.URL.Query().Get("location"); location != "" {
			query = query.Where("location LIKE ?", "%"+location+"%")
		}
		if dateVal := r.URL.Query().Get("date"); dateVal != "" {
			query = query.Where("DATE(datetime) = ?", dateVal)
		}
		if overlapping := r.URL.Query().Get("overlapping"); overlapping == "true" {
			query = query.Where(`EXISTS (
				SELECT 1 FROM appointments a2
				WHERE a2.created_by_id = appointments.created_by_id
				AND a2.id != appointments.id
				AND a2.datetime > datetime(appointments.datetime, '-30 minutes')
				AND a2.datetime < datetime(appointments.datetime, '+30 minutes')
				AND a2.deleted_at IS NULL
			)`)
		}
		if sort := r.URL.Query().Get("sort"); sort != "" {
			switch sort {
			case "name", "location", "datetime", "created_at",
				"name desc", "location desc", "datetime desc", "created_at desc":
				query = query.Order(sort)
			}
		}

		var total int64
		if err := query.Count(&total).Error; err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		var results []Appointment
		err := query.Limit(pageSize).Offset((pageNum - 1) * pageSize).Order("datetime DESC").Find(&results).Error
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		numPages := int((total + int64(pageSize) - 1) / int64(pageSize))
		objectList := components.ObjectList[Appointment]{
			Items:    results,
			Number:   pageNum,
			NumPages: numPages,
			Total:    total,
		}

		ctx := context.WithValue(r.Context(), "appointments", objectList)
		queryMap := map[string]any{}
		for param, values := range r.URL.Query() {
			if len(values) > 0 && values[0] != "" {
				queryMap[param] = values[0]
			}
		}
		ctx = context.WithValue(ctx, "$get", queryMap)
		v.RenderPage(w, r.WithContext(ctx))
	})
}

func detailHandler(v views.View) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		idStr := r.PathValue("id")
		db := r.Context().Value("$db").(*gorm.DB)
		user := r.Context().Value("$user").(p_users.User)

		appointment := getAppointmentOr404(w, r, db, idStr, user)
		if appointment == nil {
			return
		}

		ctx := r.Context()

		if appointment.GenerationID != nil {
			ctx = context.WithValue(ctx, "generation_pending", true)
		}

		appointmentMap := getters.MapFromStruct(appointment)

		overlapping := appointment.GetOverlappingAppointments(db)
		if len(overlapping) > 0 {
			var names []string
			for _, o := range overlapping {
				names = append(names, fmt.Sprintf("%s (%s)", o.Name, o.Datetime.Format("Jan 02, 15:04")))
			}
			ctx = context.WithValue(ctx, "overlap_warning", fmt.Sprintf("This appointment conflicts with: %s", joinStrings(names)))
		}

		ctx = context.WithValue(ctx, "appointment", appointmentMap)
		v.RenderPage(w, r.WithContext(ctx))
	})
}

func joinStrings(ss []string) string {
	result := ""
	for i, s := range ss {
		if i > 0 {
			result += ", "
		}
		result += s
	}
	return result
}

func createHandler(v views.View) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		db := r.Context().Value("$db").(*gorm.DB)
		user := r.Context().Value("$user").(p_users.User)

		if r.Method == http.MethodGet {
			ctx := context.WithValue(r.Context(), "$in", map[string]any{})
			v.RenderPage(w, r.WithContext(ctx))
			return
		}

		values, fieldErrors, ok := v.ParseForm(w, r)
		if !ok {
			return
		}
		if views.HasErrors(fieldErrors) {
			v.RenderWithErrors(w, r, fieldErrors, values)
			return
		}

		appointment := Appointment{
			CreatedByID: user.ID,
		}
		applyFormValues(&appointment, values)

		if err := db.Create(&appointment).Error; err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		redirectTo(w, r, fmt.Sprintf(AppUrl+"%d/", appointment.ID))
	})
}

func updateHandler(v views.View) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		idStr := r.PathValue("id")
		db := r.Context().Value("$db").(*gorm.DB)
		user := r.Context().Value("$user").(p_users.User)

		appointment := getAppointmentOr404(w, r, db, idStr, user)
		if appointment == nil {
			return
		}

		if r.Method == http.MethodGet {
			appointmentMap := getters.MapFromStruct(appointment)
			ctx := context.WithValue(r.Context(), "$in", appointmentMap)
			ctx = context.WithValue(ctx, "appointment", appointmentMap)
			v.RenderPage(w, r.WithContext(ctx))
			return
		}

		values, fieldErrors, ok := v.ParseForm(w, r)
		if !ok {
			return
		}
		if views.HasErrors(fieldErrors) {
			v.RenderWithErrors(w, r, fieldErrors, values)
			return
		}

		applyFormValues(appointment, values)

		if err := db.Save(appointment).Error; err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		redirectTo(w, r, fmt.Sprintf(AppUrl+"%d/", appointment.ID))
	})
}

func deleteHandler(v views.View) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		idStr := r.PathValue("id")
		db := r.Context().Value("$db").(*gorm.DB)
		user := r.Context().Value("$user").(p_users.User)

		appointment := getAppointmentOr404(w, r, db, idStr, user)
		if appointment == nil {
			return
		}

		if r.Method == http.MethodGet {
			ctx := context.WithValue(r.Context(), "appointment", getters.MapFromStruct(appointment))
			v.RenderPage(w, r.WithContext(ctx))
			return
		}

		if err := db.Delete(appointment).Error; err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		redirectTo(w, r, AppUrl)
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

		appointment := getAppointmentOr404(w, r, db, idStr, user)
		if appointment == nil {
			return
		}

		content := buildLetterContent(db, appointment, user.Name)
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
		user := r.Context().Value("$user").(p_users.User)

		appointment := getAppointmentOr404(w, r, db, idStr, user)
		if appointment == nil {
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
		user := r.Context().Value("$user").(p_users.User)

		appointment := getAppointmentOr404(w, r, db, idStr, user)
		if appointment == nil {
			return
		}

		ctx := context.WithValue(r.Context(), "appointment", getters.MapFromStruct(appointment))
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
		user := r.Context().Value("$user").(p_users.User)

		appointment := getAppointmentOr404(w, r, db, idStr, user)
		if appointment == nil {
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

func init() {
	lago.RegistryView.Register("appointments.ListView", p_users.AuthMiddleware(
		lago.GetPageView("appointments.AppointmentTable").WithMethod(http.MethodGet, listHandler)))

	lago.RegistryView.Register("appointments.DetailView", p_users.AuthMiddleware(
		lago.GetPageView("appointments.AppointmentDetail").WithMethod(http.MethodGet, detailHandler)))

	lago.RegistryView.Register("appointments.CreateView", p_users.AuthMiddleware(
		lago.GetPageView("appointments.AppointmentCreateForm").WithMethod(http.MethodGet, createHandler).WithMethod(http.MethodPost, createHandler)))

	lago.RegistryView.Register("appointments.UpdateView", p_users.AuthMiddleware(
		lago.GetPageView("appointments.AppointmentUpdateForm").WithMethod(http.MethodGet, updateHandler).WithMethod(http.MethodPost, updateHandler)))

	lago.RegistryView.Register("appointments.DeleteView", p_users.AuthMiddleware(
		lago.GetPageView("appointments.AppointmentDeleteForm").WithMethod(http.MethodGet, deleteHandler).WithMethod(http.MethodPost, deleteHandler)))

	lago.RegistryView.Register("appointments.GenerateView", p_users.AuthMiddleware(
		lago.GetPageView("appointments.AppointmentDetail").WithMethod(http.MethodPost, generateHandler)))

	lago.RegistryView.Register("appointments.CancelView", p_users.AuthMiddleware(
		lago.GetPageView("appointments.AppointmentDetail").WithMethod(http.MethodPost, cancelHandler)))

	lago.RegistryView.Register("appointments.AiEditFormView", p_users.AuthMiddleware(
		lago.GetPageView("appointments.AiEditModal").WithMethod(http.MethodGet, aiEditFormHandler)))

	lago.RegistryView.Register("appointments.AiEditView", p_users.AuthMiddleware(
		lago.GetPageView("appointments.AiEditModal").WithMethod(http.MethodPost, aiEditHandler)))
}
