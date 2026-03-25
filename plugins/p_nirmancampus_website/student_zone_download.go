package p_nirmancampus_website

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/views"
	"gorm.io/gorm"
)

func isHtmxRequest(r *http.Request) bool {
	return r.Header.Get("HX-Request") == "true"
}

func htmxRedirect(w http.ResponseWriter, r *http.Request, url string, code int) {
	if isHtmxRequest(r) {
		w.Header().Set("HX-Redirect", url)
		w.WriteHeader(http.StatusOK)
		return
	}
	http.Redirect(w, r, url, code)
}

func studentZoneItemHandler(_ *views.View) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		idStr := r.PathValue("id")
		id, err := strconv.ParseUint(idStr, 10, 64)
		if err != nil {
			http.NotFound(w, r)
			return
		}

		db, ok := r.Context().Value("$db").(*gorm.DB)
		if !ok || db == nil {
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		var item StudentZoneItem
		if err := db.Preload("File").First(&item, id).Error; err != nil {
			http.NotFound(w, r)
			return
		}

		if item.IsLink {
			htmxRedirect(w, r, item.Link, http.StatusFound)
			return
		}

		if item.File == nil {
			http.NotFound(w, r)
			return
		}

		// For htmx requests, redirect back to the same URL so the browser
		// performs a normal (non-boosted) GET that returns the file content.
		if isHtmxRequest(r) {
			w.Header().Set("HX-Redirect", r.URL.String())
			w.WriteHeader(http.StatusOK)
			return
		}

		download, err := item.File.OpenDownload()
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		defer download.Reader.Close()

		w.Header().Set("Content-Type", download.ContentType)
		w.Header().Set("Content-Length", fmt.Sprintf("%d", download.Size))
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", download.Filename))
		if _, err := io.Copy(w, download.Reader); err != nil {
			slog.Error("nirmancampus_website: failed writing student zone download", "id", item.ID, "error", err)
		}
	})
}

func init() {
	_ = lago.RegistryRoute.Register("nirmancampus_website.StudentZoneItemRoute", lago.Route{
		Path:    "/students-zone/item/{id}/",
		Handler: lago.NewDynamicView("nirmancampus_website.StudentZoneItemView"),
	})

	lago.RegistryView.Register("nirmancampus_website.StudentZoneItemView", &views.View{
		Handlers: map[string]func(*views.View) http.Handler{
			http.MethodGet: studentZoneItemHandler,
		},
	})
}
