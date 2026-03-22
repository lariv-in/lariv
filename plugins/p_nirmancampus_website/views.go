package p_nirmancampus_website

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/views"
	"gorm.io/gorm"
)

func parsePopupImageID(r *http.Request) (uint, error) {
	id, err := strconv.ParseUint(r.PathValue("id"), 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid popup image id")
	}
	return uint(id), nil
}

func popupImageHandler(_ *views.View) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		db, err := homePageDB(r.Context())
		if err != nil {
			slog.Error("nirmancampus_website: missing db while serving popup image", "error", err)
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		id, err := parsePopupImageID(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		node, err := loadPublicPopupImageNodeByID(db, id)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				http.NotFound(w, r)
				return
			}
			slog.Error("nirmancampus_website: failed loading popup image node", "id", id, "error", err)
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		download, err := node.OpenDownload()
		if err != nil {
			http.NotFound(w, r)
			return
		}
		defer download.Reader.Close()

		w.Header().Set("Content-Type", download.ContentType)
		w.Header().Set("Content-Length", fmt.Sprintf("%d", download.Size))
		w.Header().Set("Content-Disposition", fmt.Sprintf("inline; filename=%q", download.Filename))
		if _, err := io.Copy(w, download.Reader); err != nil {
			slog.Error("nirmancampus_website: failed writing popup image response", "id", id, "error", err)
		}
	})
}

func init() {
	lago.RegistryView.Register("nirmancampus_website.HomeView",
		lago.GetPageView("nirmancampus_website.HomePage"))

	coursesBase := lago.GetPageView("nirmancampus_website.CoursesPage")
	coursesView := &views.View{
		PageName:      coursesBase.PageName,
		PageLookup:    coursesBase.PageLookup,
		FormPatchers:  coursesBase.FormPatchers,
		QueryPatchers: coursesBase.QueryPatchers,
		Middlewares:   coursesBase.Middlewares,
		Handlers: map[string]func(*views.View) http.Handler{
			http.MethodGet: func(v *views.View) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					level := strings.TrimSpace(r.URL.Query().Get("level"))
					ctx := context.WithValue(r.Context(), "$get", map[string]any{"Level": level})
					v.RenderPage(w, r.WithContext(ctx))
				})
			},
		},
	}
	lago.RegistryView.Register("nirmancampus_website.CoursesView", coursesView)
	lago.RegistryView.Register("nirmancampus_website.PopupImageView", &views.View{
		Handlers: map[string]func(*views.View) http.Handler{
			http.MethodGet: popupImageHandler,
		},
	})
}
