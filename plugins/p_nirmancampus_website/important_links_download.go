package p_nirmancampus_website

import (
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/views"
	"gorm.io/gorm"
)

func importantLinkItemHandler(_ *views.View) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		idStr := r.PathValue("id")
		id, err := strconv.ParseUint(idStr, 10, 64)
		if err != nil {
			http.NotFound(w, r)
			return
		}

		db, dberr := getters.DBFromContext(r.Context())
		if dberr != nil {
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		item, err := gorm.G[ImportantLink](db).Preload("File", nil).Where("id = ?", id).First(r.Context())
		if err != nil {
			http.NotFound(w, r)
			return
		}

		if item.IsLink {
			views.HtmxRedirect(w, r, item.Link, http.StatusFound)
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
			// Mirrors student-zone behavior: keep serving headers even if body copy fails.
			// (Error logging is handled elsewhere in the app runtime.)
		}
	})
}

func init() {
	_ = lago.RegistryRoute.Register("nirmancampus_website.ImportantLinkItemRoute", lago.Route{
		Path:    importantLinkItemBasePath + "{id}/",
		Handler: lago.NewDynamicView("nirmancampus_website.ImportantLinkItemView"),
	})

	lago.RegistryView.Register("nirmancampus_website.ImportantLinkItemView", websiteGETOnlyView(importantLinkItemHandler))
}
