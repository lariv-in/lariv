package p_nirmancampus_website

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/views"
)

const websiteStaticPrefix = "/nirman/static/"

func websiteStaticPath(path string) string {
	cleaned := strings.TrimLeft(strings.TrimSpace(path), "/")
	return websiteStaticPrefix + cleaned
}

func resolveWebsiteStaticDir() (string, error) {
	if Config.StaticDir == "" {
		return "", fmt.Errorf("staticDir not configured")
	}

	dir := Config.StaticDir
	if !filepath.IsAbs(dir) {
		exe, err := os.Executable()
		if err != nil {
			return "", err
		}
		dir = filepath.Join(filepath.Dir(exe), dir)
	}

	st, err := os.Stat(dir)
	if err != nil {
		return "", err
	}
	if !st.IsDir() {
		return "", fmt.Errorf("staticDir is not a directory")
	}
	return dir, nil
}

func websiteStaticHandler(_ *views.View) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		dir, err := resolveWebsiteStaticDir()
		if err != nil {
			slog.Warn("nirmancampus_website: static dir unavailable", "error", err, "path", r.URL.Path)
			http.NotFound(w, r)
			return
		}

		fs := http.FileServer(http.Dir(dir))
		http.StripPrefix(websiteStaticPrefix, fs).ServeHTTP(w, r)
	})
}

func init() {
	lago.RegistryView.Register("nirmancampus_website.StaticView", &views.View{
		Handlers: map[string]func(*views.View) http.Handler{
			http.MethodGet: websiteStaticHandler,
		},
	})
}
