package p_website

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"strings"

	"github.com/lariv-in/lariv"
	"github.com/lariv-in/lariv/getters"
	"github.com/lariv-in/lariv/views"
)

type builderProjectPayload struct {
	Data any    `json:"data"`
	HTML string `json:"html"`
	CSS  string `json:"css"`
}

type builderLoadResponse struct {
	Data any    `json:"data"`
	HTML string `json:"html,omitempty"`
}

func buildPublishedHTML(html, css string) string {
	html = strings.TrimSpace(html)
	css = strings.TrimSpace(css)

	styleBlock := ""
	if css != "" {
		styleBlock = "<style>\n" + css + "\n</style>\n"
	}

	lower := strings.ToLower(html)
	if strings.Contains(lower, "<html") {
		if idx := strings.Index(lower, "</head>"); idx >= 0 {
			return html[:idx] + styleBlock + html[idx:]
		}
		if idx := strings.Index(lower, "<body"); idx >= 0 {
			return html[:idx] + "<head>\n" + styleBlock + "</head>\n" + html[idx:]
		}
		return styleBlock + html
	}

	return "<!DOCTYPE html>\n<html>\n<head>\n<meta charset=\"utf-8\">\n" +
		styleBlock +
		"</head>\n<body>\n" + html + "\n</body>\n</html>\n"
}

func buildPublishedHTMLWithTheme(html, css, themeID string, theme lariv.GrapesJSTheme) string {
	return injectThemeAssets(buildPublishedHTML(html, css), themeID, theme)
}

func loadBuilderProjectHandler(_ *views.View) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		route, ok := r.Context().Value("dbroute").(DBRoute)
		if !ok {
			http.Error(w, "route not found", http.StatusNotFound)
			return
		}

		resp := builderLoadResponse{}
		if strings.TrimSpace(route.GrapesProject) != "" {
			var project any
			if err := json.Unmarshal([]byte(route.GrapesProject), &project); err != nil {
				slog.Error("builder load: invalid grapes project JSON", "error", err, "route_id", route.ID)
				http.Error(w, "invalid stored project", http.StatusInternalServerError)
				return
			}
			resp.Data = project
		} else {
			page := route.Page
			if page.ID == 0 {
				db, err := getters.DBFromContext(r.Context())
				if err != nil {
					http.Error(w, "internal error", http.StatusInternalServerError)
					return
				}
				if err := db.First(&page, route.PageID).Error; err != nil {
					http.Error(w, "page not found", http.StatusNotFound)
					return
				}
			}
			download, err := page.OpenDownload()
			if err != nil {
				slog.Error("builder load: open page", "error", err, "page_id", page.ID)
				http.Error(w, "failed to read page", http.StatusInternalServerError)
				return
			}
			defer download.Reader.Close()
			body, err := io.ReadAll(download.Reader)
			if err != nil {
				http.Error(w, "failed to read page", http.StatusInternalServerError)
				return
			}
			resp.Data = nil
			resp.HTML = string(body)
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			slog.Error("builder load: encode response", "error", err)
		}
	})
}

func storeBuilderProjectHandler(_ *views.View) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		route, ok := r.Context().Value("dbroute").(DBRoute)
		if !ok {
			http.Error(w, "route not found", http.StatusNotFound)
			return
		}

		var payload builderProjectPayload
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			http.Error(w, "invalid JSON body", http.StatusBadRequest)
			return
		}
		if payload.Data == nil {
			http.Error(w, "project data is required", http.StatusBadRequest)
			return
		}

		projectBytes, err := json.Marshal(payload.Data)
		if err != nil {
			http.Error(w, "failed to encode project", http.StatusInternalServerError)
			return
		}

		db, err := getters.DBFromContext(r.Context())
		if err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}

		// Refresh theme from DB in case it was changed in the builder bar.
		var themeID string
		_ = db.Model(&DBRoute{}).Select("theme").Where("id = ?", route.ID).Scan(&themeID)
		if themeID == "" {
			themeID = route.Theme
		}
		var theme lariv.GrapesJSTheme
		if app, ok := lariv.AppFromContext(r.Context()); ok {
			theme, _ = lookupGrapesJSTheme(app, themeID)
		}

		published := injectDotLottieScript(buildPublishedHTMLWithTheme(payload.HTML, payload.CSS, themeID, theme))

		page := route.Page
		if page.ID == 0 {
			if err := db.First(&page, route.PageID).Error; err != nil {
				http.Error(w, "page not found", http.StatusNotFound)
				return
			}
		}

		if err := page.ReplaceContentFromReader(db, strings.NewReader(published)); err != nil {
			slog.Error("builder store: replace page content", "error", err, "page_id", page.ID)
			http.Error(w, "failed to save page content", http.StatusInternalServerError)
			return
		}

		if err := db.Model(&DBRoute{}).Where("id = ?", route.ID).Update("grapes_project", string(projectBytes)).Error; err != nil {
			slog.Error("builder store: save grapes project", "error", err, "route_id", route.ID)
			http.Error(w, "failed to save project", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"ok":true}`)
	})
}

type builderThemePayload struct {
	Theme string `json:"theme"`
}

func storeBuilderThemeHandler(_ *views.View) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		route, ok := r.Context().Value("dbroute").(DBRoute)
		if !ok {
			http.Error(w, "route not found", http.StatusNotFound)
			return
		}

		var payload builderThemePayload
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			http.Error(w, "invalid JSON body", http.StatusBadRequest)
			return
		}
		themeID := strings.TrimSpace(payload.Theme)

		if themeID != "" {
			app, ok := lariv.AppFromContext(r.Context())
			if !ok || app == nil {
				http.Error(w, "internal error", http.StatusInternalServerError)
				return
			}
			if _, found := lookupGrapesJSTheme(app, themeID); !found {
				http.Error(w, "unknown theme", http.StatusBadRequest)
				return
			}
		}

		db, err := getters.DBFromContext(r.Context())
		if err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		if err := db.Model(&DBRoute{}).Where("id = ?", route.ID).Update("theme", themeID).Error; err != nil {
			slog.Error("builder theme: save", "error", err, "route_id", route.ID)
			http.Error(w, "failed to save theme", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"ok":true}`)
	})
}
