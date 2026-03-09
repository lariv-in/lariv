package lago

import (
	"log"
	"log/slog"
	"maps"
	"net/http"

	"github.com/lariv-in/components"
	"github.com/lariv-in/views"
	"gorm.io/gorm"
)

func GetPageView(pageName string) views.View {
	_, pageExists := RegistryPage.Get(pageName)
	if !pageExists {
		log.Panicf("Tried to access page: %s, which does not exist in the template registry at this time", pageName)
	}
	return views.View{
		PageName: pageName,
		Registry: RegistryPage.All(),
		Handlers: map[string]func(v views.View) http.Handler{
			http.MethodGet: func(v views.View) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					v.RenderPage(w, r)
				})
			},
		},
	}
}

func GetModelFormCreateView(pageName string, table string) views.View {
	return GetPageView(pageName).WithMethod(http.MethodPost, func(v views.View) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			page, _ := v.GetPage()
			forms := components.FindChildren[components.FormComponent](page.(components.ParentInterface))
			if len(forms) == 0 {
				slog.Error("Could not find the form component", "view", pageName)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
			form := forms[0]
			db := r.Context().Value("$db").(*gorm.DB)
			values, errors, err := form.ParseForm(r)
			if len(errors) > 0 {
				return
			}
			if err != nil {
				return
			}
			err = db.Table(table).Create(values).Error
			if err != nil {
				slog.Error("Error while parsing form from request", "request", *r)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}

		})
	})
}

func GetModelFormUpdateView(pageName string, table string) views.View {
	return GetPageView(pageName).WithMethod(http.MethodPost, func(v views.View) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			page, _ := v.GetPage()
			forms := components.FindChildren[components.FormComponent](page.(components.ParentInterface))
			if len(forms) == 0 {
				slog.Error("Could not find the form component", "view", pageName)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
			form := forms[0]
			id := r.PathValue("id")
			db := r.Context().Value("$db").(*gorm.DB)
			instance := map[string]any{}
			err := db.Table(table).Select("ID = ?", id).First(&instance).Error
			if err != nil {
				return
			}
			values, errors, err := form.ParseForm(r)
			if len(errors) > 0 {
				return
			}
			if err != nil {
				return
			}
			maps.Copy(instance, values)
			err = db.Table(table).Save(&instance).Error
			if err != nil {
				slog.Error("Error while parsing form from request", "request", *r)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}

		})
	})
}
